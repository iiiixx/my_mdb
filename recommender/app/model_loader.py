from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import pickle
import numpy as np


@dataclass(frozen=True)
class SVDModel:
    mu: float
    bu: np.ndarray 
    bi: np.ndarray  
    pu: np.ndarray  
    qi: np.ndarray  

    raw2inner_user: dict[str, int]
    raw2inner_item: dict[str, int]
    inner2raw_user: list[str]
    inner2raw_item: list[str]

    n_users: int
    n_items: int
    k: int


def load_model(model_dir: str | Path) -> SVDModel:
    model_dir = Path(model_dir)

    ready = model_dir / "ready"
    if not ready.exists():
        raise RuntimeError(f"Model not ready (missing {ready})")

    factors_path = model_dir / "svd_factors.npz" 
    meta_path = model_dir / "svd_meta.pkl" 

    if not factors_path.exists():
        raise FileNotFoundError(f"Missing factors file: {factors_path}")
    if not meta_path.exists():
        raise FileNotFoundError(f"Missing meta file: {meta_path}")

    factors = np.load(factors_path)

    required_keys = {"mu", "bu", "bi", "pu", "qi"}
    missing = required_keys - set(factors.files)
    if missing:
        raise ValueError(f"Missing keys in factors: {sorted(missing)}")

    with open(meta_path, "rb") as f:
        meta = pickle.load(f)

    for k in ("raw2inner_user", "raw2inner_item", "inner2raw_user", "inner2raw_item"):
        if k not in meta:
            raise ValueError(f"Missing key in meta: {k}")

    mu = float(factors["mu"][0])
    bu = factors["bu"]
    bi = factors["bi"]
    pu = factors["pu"]
    qi = factors["qi"]

    if pu.ndim != 2 or qi.ndim != 2:
        raise ValueError("pu and qi must be 2D arrays")
    if bu.ndim != 1 or bi.ndim != 1:
        raise ValueError("bu and bi must be 1D arrays")
    if pu.shape[1] != qi.shape[1]:
        raise ValueError(f"Factor dim mismatch: pu={pu.shape}, qi={qi.shape}")

    n_users, k_dim = pu.shape
    n_items = qi.shape[0]

    if bu.shape[0] != n_users:
        raise ValueError(f"bu size mismatch: bu={bu.shape[0]} n_users={n_users}")
    if bi.shape[0] != n_items:
        raise ValueError(f"bi size mismatch: bi={bi.shape[0]} n_items={n_items}")

    inner2raw_user = meta["inner2raw_user"]
    inner2raw_item = meta["inner2raw_item"]

    if len(inner2raw_user) != n_users:
        raise ValueError(f"inner2raw_user size mismatch: {len(inner2raw_user)} != {n_users}")
    if len(inner2raw_item) != n_items:
        raise ValueError(f"inner2raw_item size mismatch: {len(inner2raw_item)} != {n_items}")

    return SVDModel(
        mu=mu,
        bu=bu,
        bi=bi,
        pu=pu,
        qi=qi,
        raw2inner_user=meta["raw2inner_user"],
        raw2inner_item=meta["raw2inner_item"],
        inner2raw_user=inner2raw_user,
        inner2raw_item=inner2raw_item,
        n_users=n_users,
        n_items=n_items,
        k=k_dim,
    )