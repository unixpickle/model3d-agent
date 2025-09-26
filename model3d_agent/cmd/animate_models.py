import argparse
import io
import math
import os

import numpy as np
from PIL import Image

from model3d_agent.execution import CodeExecutor


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--code_paths", type=str, nargs="+", required=True)
    parser.add_argument("--output_path", type=str, required=True)
    parser.add_argument("--frames", type=int, default=24)
    parser.add_argument("--resolution", type=int, default=512)
    parser.add_argument("--zoom", type=float, default=1.3)
    args = parser.parse_args()

    os.makedirs(args.output_path, exist_ok=True)

    code_paths = args.code_paths
    model_code = []
    for path in code_paths:
        with open(path, "r") as f:
            model_code.append(f.read())
    grid_side = int(math.sqrt(len(model_code)))
    assert grid_side**2 == len(model_code), "must pass a square number of --code_paths"

    pans = []
    executor = CodeExecutor()
    executor.warmup()
    try:
        for i, model in enumerate(model_code):
            print(f"rendering model {i}")
            pans.append(
                executor.execute_code_pan(
                    model,
                    frames=args.frames,
                    resolution=args.resolution,
                    zoom=args.zoom,
                    timeout=60 * 5,
                )
            )
    finally:
        executor.cleanup()

    print("exporting frames...")
    for frame in range(args.frames):
        imgs = [np.array(Image.open(io.BytesIO(x[frame]))) for x in pans]
        joined = np.moveaxis(
            np.stack(imgs, axis=0).reshape(
                grid_side, grid_side, args.resolution, args.resolution, 3
            ),
            1,
            2,
        ).reshape(grid_side * args.resolution, grid_side * args.resolution, 3)
        Image.fromarray(joined).save(os.path.join(args.output_path, f"{frame:03}.png"))


if __name__ == "__main__":
    main()
