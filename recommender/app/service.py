from __future__ import annotations

import signal
import grpc
from concurrent import futures
import logging

from app.model_loader import load_model
from app.recommender import Recommender
from app.logger import setup_logger
from gen import recs_pb2, recs_pb2_grpc

MAX_LIMIT = 200
class RecommenderService(recs_pb2_grpc.RecommenderServicer):
    def __init__(self, rec: Recommender, logger: logging.Logger):
        self.rec = rec
        self.log = logger

    def Recommend(self, request, context):
        user_id = int(request.user_id)
        limit = int(request.limit)

        self.log.info(f"recommend request user_id={user_id} limit={limit}")

        if limit <= 0:
            return recs_pb2.RecommendResponse(items=[])

        limit = min(limit, MAX_LIMIT)
        exclude = list(request.exclude_movie_ids)

        try:
            items = self.rec.recommend(
                user_id=user_id,
                limit=limit,
                exclude_movie_ids=exclude,
            )
        except Exception:
            self.log.exception("recommend failed")
            context.abort(grpc.StatusCode.INTERNAL, "internal error")

        return recs_pb2.RecommendResponse(
            items=[
                recs_pb2.RecommendItem(movie_id=it.movie_id, score=float(it.score))
                for it in items
            ]
        )

    def SimilarMovies(self, request, context):
        movie_id = int(request.movie_id)
        limit = int(request.limit)

        self.log.info(f"similar request movie_id={movie_id} limit={limit}")

        if limit <= 0:
            return recs_pb2.SimilarMoviesResponse(items=[])

        limit = min(limit, MAX_LIMIT)
        exclude = list(request.exclude_movie_ids)

        try:
            items = self.rec.similar_movies(
                movie_id=movie_id,
                limit=limit,
                exclude_movie_ids=exclude,
            )
        except Exception:
            self.log.exception("similar movies failed")
            context.abort(grpc.StatusCode.INTERNAL, "internal error")

        return recs_pb2.SimilarMoviesResponse(
            items=[
                recs_pb2.SimilarMovieItem(
                    movie_id=it.movie_id,
                    similarity=float(it.similarity),
                )
                for it in items
            ]
        )


def serve(
    host: str,
    port: int,
    model_dir: str,
    max_workers: int,
) -> None:
    logger = setup_logger()

    logger.info("loading model...")
    m = load_model(model_dir)
    rec = Recommender(m)

    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=max_workers),
    )

    recs_pb2_grpc.add_RecommenderServicer_to_server(
        RecommenderService(rec, logger),
        server,
    )

    addr = f"{host}:{port}"
    server.add_insecure_port(addr)
    server.start()

    logger.info(
        f"server started addr={addr} users={m.n_users} items={m.n_items} k={m.k}"
    )

    def _handle_sig(_sig, _frame):
        logger.info("shutting down server...")
        server.stop(grace=3)

    signal.signal(signal.SIGINT, _handle_sig)
    signal.signal(signal.SIGTERM, _handle_sig)

    server.wait_for_termination()