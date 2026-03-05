import os
import pickle
from datetime import datetime
from pathlib import Path

import numpy as np
import pandas as pd
from sqlalchemy import create_engine, text
from surprise import Dataset, Reader, SVD

DB_URL = os.environ.get(
    "DB_URL",
    "postgresql+psycopg://postgres:postgres@localhost:5433/mdb",
)

SAMPLE_PERCENT = int(os.environ.get("SAMPLE_PERCENT", "100"))

K = int(os.environ.get("K", "50"))

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
MODEL_ROOT = Path(BASE_DIR) / "model"
VERSIONS_DIR = MODEL_ROOT / "versions"
CURRENT_LINK = MODEL_ROOT / "current"


def load_ratings_from_db(db_url: str) -> pd.DataFrame:
    print("Step 1/4: Loading ratings from Postgres...")

    engine = create_engine(db_url, pool_pre_ping=True)

    if SAMPLE_PERCENT <= 0 or SAMPLE_PERCENT > 100:
        raise ValueError("SAMPLE_PERCENT must be in [1..100]")

    if SAMPLE_PERCENT == 100:
        sql = text("""
            SELECT user_id, movie_id, rating
            FROM ratings
            WHERE rating BETWEEN 0.5 AND 5.0
        """)
    else:
        sql = text(f"""
            SELECT user_id, movie_id, rating
            FROM ratings TABLESAMPLE SYSTEM ({SAMPLE_PERCENT})
            WHERE rating BETWEEN 0.5 AND 5.0
        """)

    df = pd.read_sql(sql, engine)

    df = df.dropna(subset=["user_id", "movie_id", "rating"])

    df["user_id"] = df["user_id"].astype(np.int32)
    df["movie_id"] = df["movie_id"].astype(np.int32)
    df["rating"] = df["rating"].astype(np.float32)

    print(f"Loaded: {len(df):,} ratings (sample={SAMPLE_PERCENT}%)")
    print("Memory (MB):", df.memory_usage(deep=True).sum() / 1024 / 1024)
    return df


def train_model(ratings: pd.DataFrame):
    print("Step 2/4: Building Surprise trainset...")

    df_surp = pd.DataFrame({
        "user_id": ratings["user_id"].astype(str),
        "movie_id": ratings["movie_id"].astype(str),
        "rating": ratings["rating"].astype(float), 
    })

    reader = Reader(rating_scale=(0.5, 5.0))
    data = Dataset.load_from_df(df_surp[["user_id", "movie_id", "rating"]], reader)
    trainset = data.build_full_trainset()

    print(f"Step 3/4: Training SVD (k={K})...")
    model = SVD(n_factors=K, random_state=42)
    model.fit(trainset)

    return model, trainset


def build_mappings(trainset):
    print("Step 4/4: Building ID mappings...")

    raw2inner_user = dict(trainset._raw2inner_id_users)
    raw2inner_item = dict(trainset._raw2inner_id_items)

    inner2raw_user = [None] * trainset.n_users
    for raw, inner in raw2inner_user.items():
        inner2raw_user[inner] = raw

    inner2raw_item = [None] * trainset.n_items
    for raw, inner in raw2inner_item.items():
        inner2raw_item[inner] = raw

    return {
        "raw2inner_user": raw2inner_user,
        "raw2inner_item": raw2inner_item,
        "inner2raw_user": inner2raw_user,
        "inner2raw_item": inner2raw_item,
    }


def save_light_versioned(model, trainset, mappings):
    print("Saving lightweight model files (versioned)...")

    MODEL_ROOT.mkdir(parents=True, exist_ok=True)
    VERSIONS_DIR.mkdir(parents=True, exist_ok=True)

    ver = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    out_dir = VERSIONS_DIR / ver
    out_dir.mkdir(parents=True, exist_ok=False)

    factors_path = out_dir / "svd_factors.npz"
    meta_path = out_dir / "svd_meta.pkl"

    mu = float(trainset.global_mean)

    np.savez_compressed(
        factors_path,
        mu=np.array([mu], dtype=np.float32),
        bu=model.bu.astype(np.float32),
        bi=model.bi.astype(np.float32),
        pu=model.pu.astype(np.float32),
        qi=model.qi.astype(np.float32),
    )

    with open(meta_path, "wb") as f:
        pickle.dump(mappings, f, protocol=pickle.HIGHEST_PROTOCOL)

    (out_dir / "ready").write_text("ok")

    tmp_link = MODEL_ROOT / ".current_tmp"
    if tmp_link.exists() or tmp_link.is_symlink():
        tmp_link.unlink()
    tmp_link.symlink_to(out_dir)
    os.replace(tmp_link, CURRENT_LINK)

    print("Activated model:", out_dir)
    print("Factors size (MB):", factors_path.stat().st_size / 1024 / 1024)
    print("Meta size (MB):", meta_path.stat().st_size / 1024 / 1024)

def cleanup_old_models(keep=3):
    versions = sorted(VERSIONS_DIR.iterdir(), reverse=True)

    for old in versions[keep:]:
        print("Removing old model:", old)
        import shutil
        shutil.rmtree(old, ignore_errors=True)

def main():
    print("DB_URL:", DB_URL)
    print("SAMPLE_PERCENT:", SAMPLE_PERCENT)
    print("K:", K)

    ratings = load_ratings_from_db(DB_URL)
    model, trainset = train_model(ratings)
    mappings = build_mappings(trainset)
    save_light_versioned(model, trainset, mappings)
    cleanup_old_models(keep=5)


if __name__ == "__main__":
    main()