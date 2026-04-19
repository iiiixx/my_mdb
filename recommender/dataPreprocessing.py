import pandas as pd

print(f"\n# Этап 1. Загрузка данных")
ratings = pd.read_csv("/Library/go_projects/my_mdb/init_db/csv/rating.csv")
movies = pd.read_csv("/Library/go_projects/my_mdb/init_db/csv/movie.csv")
links = pd.read_csv("/Library/go_projects/my_mdb/init_db/csv/link.csv")
tags = pd.read_csv("/Library/go_projects/my_mdb/init_db/csv/tag.csv")
genome_tags = pd.read_csv("/Library/go_projects/my_mdb/init_db/csv/genome_tags.csv")
genome_scores = pd.read_csv("/Library/go_projects/my_mdb/init_db/csv/genome_scores.csv")

'''
print(f"\n# Этап 2. Проверка размеров таблиц и структуры")
tables = {
    "ratings": ratings,
    "movies": movies,
    "links": links,
    "tags": tags,
    "genome_tags": genome_tags,
    "genome_scores": genome_scores,
}

for name, df in tables.items():
    print(f"\n{name}")
    print("shape:", df.shape)
    print(df.dtypes)

print(f"\n# Этап 3. Проверка пропущенных значений")
for name, df in tables.items():
    print(f"\Пропущенные значения в {name}:")
    print(df.isnull().sum())

print(f"\n# Этап 4. Проверка дубликатов")
for name, df in tables.items():
    dup_count = df.duplicated().sum()
    print(f"{name}: строки с дубликатами = {dup_count}")

pair_duplicates = ratings.duplicated(subset=["userId", "movieId"]).sum()
print("Пары userId-movieId с дубликатами:", pair_duplicates)

ratings = ratings.sort_values("timestamp")
ratings = ratings.drop_duplicates(subset=["userId", "movieId"], keep="last")

print(f"\n# Этап 5. Проверка диапазона рейтингов")
print("Минимальный рейтинг:", ratings["rating"].min())
print("Максимальный рейтинг:", ratings["rating"].max())
print("Уникальные рейтинги:", sorted(ratings["rating"].unique())[:20])

valid_scale = set([x / 2 for x in range(1, 11)])  # 0.5 ... 5.0
invalid_ratings = ratings[~ratings["rating"].isin(valid_scale)]
print("Некорректные рейтинги:", len(invalid_ratings))

print(f"\n# Этап 6. Базовая статистика по данным для модели")
n_users = ratings["userId"].nunique()
n_movies = ratings["movieId"].nunique()
n_ratings = len(ratings)

print("Users:", n_users)
print("Movies rated:", n_movies)
print("Ratings:", n_ratings)

print(f"\n# Этап 7. Анализ активности пользователей и фильмов")
ratings_per_user = ratings.groupby("userId").size()
print(ratings_per_user.describe())

ratings_per_movie = ratings.groupby("movieId").size()
print(ratings_per_movie.describe())

print(f"\n# Этап 8. Оценка разреженности матрицы")

num_possible = n_users * n_movies
sparsity = 1 - (n_ratings / num_possible)

print("Possible interactions:", num_possible)
print("Observed interactions:", n_ratings)
print("Sparsity:", sparsity)
'''

'''
import math
import random
from collections import defaultdict

import numpy as np
import pandas as pd
from surprise import Dataset, Reader, SVD, accuracy

# =========================================================
# Этап 9. Подготовка данных для модели
# =========================================================

model_data = ratings[["userId", "movieId", "rating"]].copy()

# для Surprise удобнее строковые идентификаторы
model_data["userId"] = model_data["userId"].astype(str)
model_data["movieId"] = model_data["movieId"].astype(str)
model_data["rating"] = model_data["rating"].astype(float)

print("\n# Этап 9. Подготовка данных для модели")
print(model_data.head())
print("Model data shape:", model_data.shape)


# =========================================================
# Этап 10. Разделение на обучающую и тестовую выборки
# Пер-user split: у каждого пользователя часть рейтингов уходит в test
# =========================================================

def split_per_user(
    df: pd.DataFrame,
    test_ratio: float = 0.2,
    min_user_ratings: int = 5,
    random_state: int = 42,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    """
    Делит данные по пользователям:
    - пользователи с числом рейтингов < min_user_ratings исключаются из оценки;
    - для каждого пользователя часть рейтингов уходит в test;
    - минимум 1 рейтинг остается в train.
    """
    rng = np.random.default_rng(random_state)

    user_counts = df.groupby("userId").size()
    valid_users = user_counts[user_counts >= min_user_ratings].index

    df = df[df["userId"].isin(valid_users)].copy()

    train_parts = []
    test_parts = []

    for user_id, group in df.groupby("userId", sort=False):
        idx = group.index.to_numpy()
        rng.shuffle(idx)

        n = len(idx)
        n_test = max(1, int(round(n * test_ratio)))
        n_test = min(n_test, n - 1)  # хотя бы 1 объект оставляем в train

        test_idx = idx[:n_test]
        train_idx = idx[n_test:]

        train_parts.append(df.loc[train_idx])
        test_parts.append(df.loc[test_idx])

    train_df = pd.concat(train_parts, ignore_index=True)
    test_df = pd.concat(test_parts, ignore_index=True)

    return train_df, test_df


train_df, test_df = split_per_user(
    model_data,
    test_ratio=0.2,
    min_user_ratings=5,
    random_state=42,
)

print("\n# Этап 10. Разделение выборок")
print("Train size:", len(train_df))
print("Test size:", len(test_df))
print("Train users:", train_df["userId"].nunique())
print("Test users:", test_df["userId"].nunique())
print("Train movies:", train_df["movieId"].nunique())
print("Test movies:", test_df["movieId"].nunique())

# проверка: все пользователи test присутствуют в train
missing_users = set(test_df["userId"].unique()) - set(train_df["userId"].unique())
print("Users in test but not in train:", len(missing_users))
'''

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

# базовые размеры
n_ratings = len(ratings)
n_users = ratings["userId"].nunique()
n_movies = ratings["movieId"].nunique()

print("Количество рейтингов:", n_ratings)
print("Количество пользователей:", n_users)
print("Количество фильмов:", n_movies)

# средние показатели
avg_ratings_per_user = n_ratings / n_users
avg_ratings_per_movie = n_ratings / n_movies

print("Среднее число оценок на пользователя:", round(avg_ratings_per_user, 2))
print("Среднее число оценок на фильм:", round(avg_ratings_per_movie, 2))

# плотность и разреженность матрицы
density = n_ratings / (n_users * n_movies)
sparsity = 1 - density

print("Плотность матрицы:", round(density, 6))
print("Разреженность матрицы:", round(sparsity * 100, 4), "%")
'''
user_activity = ratings.groupby("userId")["movieId"].count()

print(user_activity.describe())

plt.figure(figsize=(10, 5))
plt.hist(user_activity[user_activity <= 500], bins=50)
plt.title("Распределение количества оценок на пользователя")
plt.xlabel("Количество оценок")
plt.ylabel("Число пользователей")
plt.grid(True, alpha=0.3)
plt.show()

movie_activity = ratings.groupby("movieId")["userId"].count()

print(movie_activity.describe())

plt.figure(figsize=(10, 5))
plt.hist(movie_activity[movie_activity <= 1000], bins=50)
plt.title("Распределение количества оценок на фильм")
plt.xlabel("Количество оценок")
plt.ylabel("Число фильмов")
plt.grid(True, alpha=0.3)
plt.show()
'''
import matplotlib.pyplot as plt
from matplotlib.ticker import FuncFormatter

rating_dist = ratings["rating"].value_counts().sort_index()
print(rating_dist)

plt.figure(figsize=(10,5))
bars = plt.bar(rating_dist.index.astype(str), rating_dist.values)

plt.title("Распределение значений пользовательских рейтингов")
plt.xlabel("Оценка")
plt.ylabel("Количество")

# Функция для форматирования значений в миллионах на оси Y
def millions(x, pos):
    return f'{x/1e6:.1f}M' if x >= 1e6 else f'{int(x)}'

plt.gca().yaxis.set_major_formatter(FuncFormatter(millions))

# Добавление подписей значений над столбцами
for bar in bars:
    height = bar.get_height()
    plt.text(bar.get_x() + bar.get_width()/2., height,
        f'{height/1e6:.1f}M', ha='center', va='bottom')


plt.grid(True, axis="y", alpha=0.3)
plt.tight_layout()
plt.show()