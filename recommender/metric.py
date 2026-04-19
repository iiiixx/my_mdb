import pandas as pd
import numpy as np
from collections import defaultdict

from surprise import Dataset, Reader, SVD
from surprise.model_selection import train_test_split


# =========================
# 1. Загрузка данных
# =========================
ratings = pd.read_csv("/Library/go_projects/my_mdb/init_db/csv/rating.csv")
ratings = ratings.sample(5000000, random_state=42)
# оставляем только нужные поля
ratings = ratings[["userId", "movieId", "rating"]].copy()

# для Surprise удобнее строковые идентификаторы
ratings["userId"] = ratings["userId"].astype(str)
ratings["movieId"] = ratings["movieId"].astype(str)
ratings["rating"] = ratings["rating"].astype(float)

print("Всего оценок:", len(ratings))
print("Пользователей:", ratings["userId"].nunique())
print("Фильмов:", ratings["movieId"].nunique())


# =========================
# 2. Подготовка данных
# =========================
reader = Reader(rating_scale=(ratings["rating"].min(), ratings["rating"].max()))
data = Dataset.load_from_df(ratings[["userId", "movieId", "rating"]], reader)

trainset, testset = train_test_split(data, test_size=0.2, random_state=42)

print("Train size:", trainset.n_ratings)
print("Test size:", len(testset))


# =========================
# 3. Обучение модели
# =========================
model = SVD(
    n_factors=100,
    n_epochs=20,
    lr_all=0.005,
    reg_all=0.02,
    random_state=42
)

model.fit(trainset)


# =========================
# 4. Предсказания на test
# =========================
predictions = model.test(testset)


# =========================
# 5. Метрики Precision@K, Recall@K
# =========================
def precision_recall_at_k(predictions, k=10, threshold=4.0):
    """
    predictions: список объектов Prediction из surprise
    k: top-k
    threshold: порог релевантности
    """
    user_est_true = defaultdict(list)

    for pred in predictions:
        uid = pred.uid
        est = pred.est
        true_r = pred.r_ui
        user_est_true[uid].append((est, true_r))

    precisions = {}
    recalls = {}

    for uid, user_ratings in user_est_true.items():
        # сортировка по предсказанному рейтингу
        user_ratings.sort(key=lambda x: x[0], reverse=True)

        top_k = user_ratings[:k]

        # сколько релевантных в test вообще
        n_rel = sum(true_r >= threshold for (_, true_r) in user_ratings)

        # сколько рекомендованных в top-k
        n_rec_k = len(top_k)

        # сколько релевантных среди рекомендованных
        n_rel_and_rec_k = sum(true_r >= threshold for (_, true_r) in top_k)

        precisions[uid] = n_rel_and_rec_k / n_rec_k if n_rec_k != 0 else 0
        recalls[uid] = n_rel_and_rec_k / n_rel if n_rel != 0 else 0

    precision = np.mean(list(precisions.values()))
    recall = np.mean(list(recalls.values()))

    return precision, recall


# =========================
# 6. Метрика NDCG@K
# =========================
def ndcg_at_k(predictions, k=10, threshold=4.0):
    user_est_true = defaultdict(list)

    for pred in predictions:
        uid = pred.uid
        est = pred.est
        true_r = pred.r_ui
        user_est_true[uid].append((est, true_r))

    ndcgs = []

    for uid, user_ratings in user_est_true.items():
        # сортируем по предсказанным оценкам
        ranked = sorted(user_ratings, key=lambda x: x[0], reverse=True)[:k]

        # DCG
        dcg = 0.0
        for i, (_, true_r) in enumerate(ranked, start=1):
            rel = 1 if true_r >= threshold else 0
            dcg += rel / np.log2(i + 1)

        # идеальное ранжирование
        ideal_rels = sorted(
            [1 if true_r >= threshold else 0 for (_, true_r) in user_ratings],
            reverse=True
        )[:k]

        idcg = 0.0
        for i, rel in enumerate(ideal_rels, start=1):
            idcg += rel / np.log2(i + 1)

        ndcg = dcg / idcg if idcg > 0 else 0.0
        ndcgs.append(ndcg)

    return np.mean(ndcgs)


# =========================
# 7. Расчёт метрик
# =========================
K = 10
THRESHOLD = 4.0

precision, recall = precision_recall_at_k(predictions, k=K, threshold=THRESHOLD)
ndcg = ndcg_at_k(predictions, k=K, threshold=THRESHOLD)

print(f"Precision@{K}: {precision:.4f}")
print(f"Recall@{K}:    {recall:.4f}")
print(f"NDCG@{K}:      {ndcg:.4f}")