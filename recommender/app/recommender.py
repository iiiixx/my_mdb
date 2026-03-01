from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable
import numpy as np

from .model_loader import SVDModel


@dataclass(frozen=True)
class RecommendItem:
    movie_id: int
    score: float


@dataclass(frozen=True)
class SimilarItem:
    movie_id: int
    similarity: float


class Recommender:
    """
    SVD + biases:
      r_hat(u,i) = mu + b_u + b_i + p_u^T q_i

    Similarity:
      sim(i,j) = cosine(q_i, q_j)
    """

    def __init__(self, model: SVDModel):
        self.m = model
        qi = np.asarray(self.m.qi, dtype=np.float32)
        norms = np.linalg.norm(qi, axis=1, keepdims=True)
        eps = 1e-12
        self._qi_norm = qi / (norms + eps)

    def recommend(
        self,
        user_id: int,
        limit: int = 20,
        exclude_movie_ids: Iterable[int] | None = None,
        min_score: float | None = None,
    ) -> list[RecommendItem]:
        if limit <= 0:
            return []

        uid = self.m.raw2inner_user.get(str(user_id))
        if uid is None:
            return []

        pu_u = self.m.pu[uid]
        scores = self.m.mu + self.m.bu[uid] + self.m.bi + (self.m.qi @ pu_u)

        if exclude_movie_ids:
            bad_inner = []
            for mid in exclude_movie_ids:
                iid = self.m.raw2inner_item.get(str(mid))
                if iid is not None:
                    bad_inner.append(iid)

            if bad_inner:
                scores = scores.copy()
                scores[np.array(bad_inner, dtype=np.int64)] = -np.inf

        if min_score is not None:
            scores = scores.copy()
            scores[scores < float(min_score)] = -np.inf

        n = min(limit, scores.shape[0])
        idx = np.argpartition(-scores, n - 1)[:n]
        idx = idx[np.argsort(-scores[idx])]

        out: list[RecommendItem] = []
        for iid in idx:
            raw_mid = self.m.inner2raw_item[int(iid)]
            out.append(RecommendItem(movie_id=int(raw_mid), score=float(scores[iid])))

        return out

    def similar_movies(
        self,
        movie_id: int,
        limit: int = 20,
        exclude_movie_ids: Iterable[int] | None = None,
        min_similarity: float | None = None,
    ) -> list[SimilarItem]:
        if limit <= 0:
            return []

        iid = self.m.raw2inner_item.get(str(movie_id))
        if iid is None:
            return []

        v = self._qi_norm[int(iid)]  
        sims = self._qi_norm @ v     

        sims = sims.copy()
        sims[int(iid)] = -np.inf

        if exclude_movie_ids:
            bad_inner = []
            for mid in exclude_movie_ids:
                j = self.m.raw2inner_item.get(str(mid))
                if j is not None:
                    bad_inner.append(j)
            if bad_inner:
                sims[np.array(bad_inner, dtype=np.int64)] = -np.inf

        if min_similarity is not None:
            thr = float(min_similarity)
            sims[sims < thr] = -np.inf

        n = min(limit, sims.shape[0])
        idx = np.argpartition(-sims, n - 1)[:n]
        idx = idx[np.argsort(-sims[idx])]

        out: list[SimilarItem] = []
        for j in idx:
            raw_mid = self.m.inner2raw_item[int(j)]
            out.append(SimilarItem(movie_id=int(raw_mid), similarity=float(sims[int(j)])))

        return out