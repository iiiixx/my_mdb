from __future__ import annotations

import os
from app.service import serve

from app.logger import setup_logger


def main() -> None:
    log_level = os.getenv("RECS_LOG_LEVEL", "INFO")
    setup_logger(log_level)
    host = os.getenv("RECS_HOST", "0.0.0.0")
    port = int(os.getenv("RECS_PORT", "50051"))
    model_dir = os.getenv("RECS_MODEL_DIR", "model")
    workers = int(os.getenv("RECS_WORKERS", "8"))

    serve(
        host=host,
        port=port,
        model_dir=model_dir,
        max_workers=workers,
    )


if __name__ == "__main__":
    main()