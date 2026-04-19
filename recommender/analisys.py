import pandas as pd
from pathlib import Path

data_dir = Path("/Library/go_projects/my_mdb/init_db/csv")

# список файлов
files = [
    "rating.csv",
    "movie.csv",
    "link.csv",
    "tag.csv",
    "genome_tags.csv",
    "genome_scores.csv"
]

for file in files:
    file_path = data_dir / file
    print(f"\n{file}")

    try:
        df = pd.read_csv(file_path)
        print(df.head(3))
    except Exception as e:
        print(f"Ошибка чтения файла: {e}")
