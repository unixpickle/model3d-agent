import io
import os
import random
import shutil
import subprocess
import tempfile
from contextlib import contextmanager
from typing import Iterator

from PIL import Image

from .sandbox import run_sandboxed


class ExecutionError(Exception):
    def __init__(self, message: str):
        super().__init__(message)
        self.message = message


class CodeExecutor:
    def __init__(self):
        self.directory = tempfile.TemporaryDirectory()
        self.warmed_up = False

    def warmup(self):
        self.warmed_up = True
        subprocess.check_output(
            ["go", "mod", "init", "codeexecutor"],
            cwd=self.directory.name,
            stderr=subprocess.STDOUT,
        )
        subprocess.check_output(
            ["go", "get", "github.com/unixpickle/model3d/model3d@v0.4.6"],
            cwd=self.directory.name,
            stderr=subprocess.STDOUT,
        )

    def cleanup(self):
        self.directory.cleanup()

    @contextmanager
    def _execution_dir(self) -> Iterator[str]:
        assert self.warmed_up, "must call warmup() first"
        with tempfile.TemporaryDirectory() as tmp_dir:
            shutil.copytree(self.directory.name, tmp_dir, dirs_exist_ok=True)
            yield tmp_dir

    def execute_code(
        self, code: str, memory_limit: int = 1_000_000_000, timeout: float = 60.0
    ) -> bytes:
        """Execute the code and return the rendered image bytes."""
        with self._execution_dir() as tmp_dir:
            tmp_image_name = f"rendering_{random.randint(0, 1000000)}.png"
            with open(os.path.join(tmp_dir, "main.go"), "w") as f:
                f.write(code + "\n\n")
                f.write(
                    f"""\
func main() {{
    mesh, colorFunc := CreateModel()
    render3d.SaveRandomGrid("{tmp_image_name}", mesh, 3, 3, 300, colorFunc.RenderColor)
}}"""
                )

            proc = subprocess.run(
                ["go", "build", "-o", "executable"],
                cwd=tmp_dir,
                capture_output=True,
                text=True,
            )
            if proc.returncode != 0:
                raise ExecutionError("Failed to build:\n\n" + proc.stderr)

            try:
                result = run_sandboxed(
                    [os.path.join(tmp_dir, "executable")],
                    tmp_dir,
                    mem_bytes=memory_limit,
                    timeout=timeout,
                )
            except subprocess.TimeoutExpired:
                raise ExecutionError("Program execution timed out.")
            if result.returncode:
                raise ExecutionError(
                    f"Program returned status: {result.returncode}\n\n"
                    + f"Standard output (may be truncated):\n{result.stdout}\n\n"
                    + f"Standard error (may be truncated):\n{result.stderr}"
                )
            img_path = os.path.join(tmp_dir, tmp_image_name)
            if not os.path.exists(img_path):
                raise ExecutionError("Program exited without saving a rendering.")
            try:
                img = Image.open(img_path)
                buf = io.BytesIO()
                img.save(buf, format="JPEG", quality=90)
                return buf.getvalue()
            except OSError:
                raise ExecutionError("PNG data is not valid.")


def run_example():
    ex = CodeExecutor()
    try:
        ex.warmup()
        out = ex.execute_code(
            """\
package main

import (
    "github.com/unixpickle/model3d/model3d"
    "github.com/unixpickle/model3d/render3d"
    "github.com/unixpickle/model3d/toolbox3d"
)

func CreateModel() (*model3d.Mesh, toolbox3d.CoordColorFunc) {
    solid := model3d.JoinedSolid{
        model3d.NewRect(model3d.XYZ(-1, -1, -1), model3d.XYZ(1, 1, 1)),
        &model3d.Sphere{Center: model3d.Z(1.5), Radius: 0.5},
    }
    delta := 0.05
    mesh := model3d.DualContour(solid, delta, true, false) // repair=true, clip=false
    return mesh, func(c model3d.Coord3D) render3d.Color {
        if c.Z > 1.001 {
            return render3d.NewColorRGB(1, 0, 0)
        } else {
            return render3d.NewColorRGB(0, 0, 1)
        }
    }
}"""
        )
        with open("example.jpg", "wb") as f:
            f.write(out)
    finally:
        ex.cleanup()


if __name__ == "__main__":
    run_example()
