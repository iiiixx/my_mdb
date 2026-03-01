import os
import pickle
import numpy as np
import pandas as pd
from pathlib import Path
from surprise import Dataset, Reader, SVD

CSV_PATH = "/Library/go_projects/my_mdb/init_db/csv/rating.csv"

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
MODEL_DIR = Path(BASE_DIR) / "model"
FACTORS_PATH = MODEL_DIR / "svd_factors.npz"
META_PATH = MODEL_DIR / "svd_meta.pkl"

K = 50


def load_ratings(path: str) -> pd.DataFrame:
    print(f"Step 1/4: Loading ratings CSV: {path}")
    df = pd.read_csv(path)

    df = df.rename(columns={
        "userId": "user_id",
        "movieId": "movie_id",
        "rating": "rating",
    })

    df = df[["user_id", "movie_id", "rating"]].copy()
    df = df.dropna()
    df = df[df["rating"].between(0.5, 5.0)]

    df["user_id"] = df["user_id"].astype(int).astype(str)
    df["movie_id"] = df["movie_id"].astype(int).astype(str)
    df["rating"] = df["rating"].astype(float)

    print(f"Loaded: {len(df):,} ratings")
    return df


def train_model(ratings: pd.DataFrame):
    print(f"Step 2/4: Building Surprise trainset...")
    reader = Reader(rating_scale=(0.5, 5.0))
    data = Dataset.load_from_df(ratings[["user_id", "movie_id", "rating"]], reader)
    trainset = data.build_full_trainset()

    print(f"Step 3/4: Training SVD (k={K})...")
    model = SVD(n_factors=K, random_state=42)
    model.fit(trainset)

    return model, trainset


def build_mappings(trainset):
    print(f"Step 4/4: Building ID mappings...")
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


def save_light(model, trainset, mappings):
    print(f"Saving lightweight model files...")
    MODEL_DIR.mkdir(parents=True, exist_ok=True)

    mu = float(trainset.global_mean)

    np.savez_compressed(
        FACTORS_PATH,
        mu=np.array([mu], dtype=np.float32),
        bu=model.bu.astype(np.float32),
        bi=model.bi.astype(np.float32),
        pu=model.pu.astype(np.float32),
        qi=model.qi.astype(np.float32),
    )

    with open(META_PATH, "wb") as f:
        pickle.dump(mappings, f, protocol=pickle.HIGHEST_PROTOCOL)

    print("Saved lightweight model.")
    print("Factors size (MB):", FACTORS_PATH.stat().st_size / 1024 / 1024)
    print("Meta size (MB):", META_PATH.stat().st_size / 1024 / 1024)


def main():
    ratings = load_ratings(CSV_PATH)
    model, trainset = train_model(ratings)
    mappings = build_mappings(trainset)
    save_light(model, trainset, mappings)


if __name__ == "__main__":
    main()