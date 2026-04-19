import os
import pandas as pd
import numpy as np
from surprise import Dataset, Reader
from surprise.model_selection import train_test_split
from sqlalchemy import create_engine

# =========================
# 1. Настройки
# =========================
DB_URL = os.environ.get(
    "DB_URL",
    "postgresql+psycopg://postgres:postgres@localhost:5433/mdb",
)

K = 10
THRESHOLD = 4.0
RATINGS_PATH = "/Library/go_projects/my_mdb/init_db/csv/rating.csv"
SAMPLE_SIZE = 5_000_000

# =========================
# 2. Загрузка ratings
# =========================
ratings = pd.read_csv(RATINGS_PATH)
ratings = ratings.sample(SAMPLE_SIZE, random_state=42)
ratings = ratings[["userId", "movieId", "rating"]].copy()

ratings["userId"] = ratings["userId"].astype(str)
ratings["movieId"] = ratings["movieId"].astype(str)
ratings["rating"] = ratings["rating"].astype(float)

print("Всего оценок:", len(ratings))
print("Пользователей:", ratings["userId"].nunique())
print("Фильмов:", ratings["movieId"].nunique())

# =========================
# 3. Train / test split
# =========================
reader = Reader(rating_scale=(ratings["rating"].min(), ratings["rating"].max()))
data = Dataset.load_from_df(ratings[["userId", "movieId", "rating"]], reader)

trainset, testset = train_test_split(data, test_size=0.2, random_state=42)

print("Train size:", trainset.n_ratings)
print("Test size:", len(testset))

# Восстанавливаем train в DataFrame
train_records = []
for uid_inner, iid_inner, r in trainset.all_ratings():
    uid = trainset.to_raw_uid(uid_inner)
    iid = trainset.to_raw_iid(iid_inner)
    train_records.append((uid, iid, float(r)))

train_df = pd.DataFrame(train_records, columns=["userId", "movieId", "rating"])
test_df = pd.DataFrame(testset, columns=["userId", "movieId", "rating"])

print("Train users:", train_df["userId"].nunique())
print("Test users:", test_df["userId"].nunique())

# =========================
# 4. Загрузка movie_rating_stats из PostgreSQL
# =========================
engine = create_engine(DB_URL)

query = """
SELECT
    movie_id,
    weighted_score,
    votes
FROM movie_rating_stats
WHERE weighted_score IS NOT NULL
ORDER BY weighted_score DESC, votes DESC, movie_id ASC
"""

movie_stats = pd.read_sql(query, engine)

movie_stats["movie_id"] = movie_stats["movie_id"].astype(str)

global_top_movies = movie_stats["movie_id"].tolist()

print("Фильмов в movie_rating_stats:", len(movie_stats))
print("Фильмов в глобальном top:", len(global_top_movies))

# =========================
# 5. Подготовка пользовательских структур
# =========================
# Фильмы, уже встречавшиеся пользователю в train
train_seen = train_df.groupby("userId")["movieId"].apply(set).to_dict()

# Релевантные фильмы из test (например rating >= 4.0)
test_relevant = (
    test_df[test_df["rating"] >= THRESHOLD]
    .groupby("userId")["movieId"]
    .apply(set)
    .to_dict()
)

users = sorted(set(test_df["userId"]))

print("Пользователей для оценки:", len(users))
print("Пользователей с релевантными фильмами в test:", len(test_relevant))

# =========================
# 6. Генерация рекомендаций baseline
# =========================
def get_popular_top_k_for_user(uid: str, global_top_movies: list[str], train_seen: dict[str, set[str]], k: int = 10) -> list[str]:
    seen = train_seen.get(uid, set())
    recs = []

    for movie_id in global_top_movies:
        if movie_id not in seen:
            recs.append(movie_id)
        if len(recs) == k:
            break

    return recs

# =========================
# 7. Метрики
# =========================
def evaluate_popular_baseline(
    users: list[str],
    global_top_movies: list[str],
    train_seen: dict[str, set[str]],
    test_relevant: dict[str, set[str]],
    k: int = 10,
) -> tuple[float, float, float]:
    precisions = []
    recalls = []
    ndcgs = []

    for uid in users:
        recs = get_popular_top_k_for_user(uid, global_top_movies, train_seen, k=k)
        relevant = test_relevant.get(uid, set())

        # Precision@K
        n_rel_and_rec_k = sum(1 for movie_id in recs if movie_id in relevant)
        precision = n_rel_and_rec_k / len(recs) if len(recs) > 0 else 0.0
        precisions.append(precision)

        # Recall@K
        n_rel = len(relevant)
        recall = n_rel_and_rec_k / n_rel if n_rel > 0 else 0.0
        recalls.append(recall)

        # NDCG@K
        dcg = 0.0
        for i, movie_id in enumerate(recs, start=1):
            rel = 1 if movie_id in relevant else 0
            dcg += rel / np.log2(i + 1)

        ideal_rels = [1] * min(len(relevant), k)
        idcg = 0.0
        for i, rel in enumerate(ideal_rels, start=1):
            idcg += rel / np.log2(i + 1)

        ndcg = dcg / idcg if idcg > 0 else 0.0
        ndcgs.append(ndcg)

    return float(np.mean(precisions)), float(np.mean(recalls)), float(np.mean(ndcgs))

# =========================
# 8. Расчёт метрик baseline
# =========================
precision_pop, recall_pop, ndcg_pop = evaluate_popular_baseline(
    users=users,
    global_top_movies=global_top_movies,
    train_seen=train_seen,
    test_relevant=test_relevant,
    k=K,
)

print(f"Popular baseline Precision@{K}: {precision_pop:.4f}")
print(f"Popular baseline Recall@{K}:    {recall_pop:.4f}")
print(f"Popular baseline NDCG@{K}:      {ndcg_pop:.4f}")