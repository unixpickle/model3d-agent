import os
import shutil
import subprocess
import sys
import tempfile
from dataclasses import dataclass
from threading import Lock, Thread

CurrentFile = os.path.abspath(__file__)


def _build_profile(allowed_dir: str) -> str:
    allowed_dir = os.path.abspath(allowed_dir)
    readable = [
        "/",
    ]

    read_forms = "\n  ".join(f'(subpath "{p}")' for p in readable)
    profile = f"""\
(version 1)
(deny default)
(allow process*)
(deny network*)
(allow sysctl-read)

;; allow reads for system/runtime + your tool locations
(allow file-read*
  {read_forms}
)

;; read/write for your sandboxes data root
(allow file-read*  (subpath "{allowed_dir}"))
(allow file-write* (subpath "{allowed_dir}"))
"""
    return profile


@dataclass
class ProcessOutput:
    returncode: int
    stdout: str
    stderr: str


def run_sandboxed(
    cmd: list[str],
    allowed_dir: str,
    mem_bytes: int,
    cpu_seconds: int | None = None,
    env: dict[str, str] | None = None,
    timeout: float | None = None,
    max_output_bytes: int = 32768,
) -> ProcessOutput:
    """
    Run `cmd` under a macOS seatbelt sandbox with FS writes restricted to
    `allowed_dir` and no network, and cap memory via RLIMITs.

    May raise a subprocess.TimeoutExpired if execution fails.
    """
    if shutil.which("sandbox-exec") is None:
        raise RuntimeError(
            "sandbox-exec not found (deprecated but still present on macOS)."
        )

    allowed_dir = os.path.realpath(os.path.abspath(allowed_dir))
    profile = _build_profile(allowed_dir)

    with tempfile.TemporaryDirectory(prefix="py-seatbelt-") as td:
        profile_path = os.path.join(td, "profile.sb")

        with open(profile_path, "w") as f:
            f.write(profile)

        # Make sure the shim directory is included as readable (already added via extra_read_paths)
        # Build the sandboxed command: sandbox-exec -f profile -- python shim mem <cmd...>
        full_cmd = [
            "sandbox-exec",
            "-f",
            profile_path,
            "--",
            sys.executable,
            CurrentFile,
            str(int(mem_bytes)),
            *cmd,
        ]

        run_env = os.environ.copy()
        if env:
            run_env.update(env)
        if cpu_seconds is not None:
            run_env["SB_CPU_SECONDS"] = str(int(cpu_seconds))

        cwd = allowed_dir

        return run_program_get_output(
            full_cmd,
            cwd=cwd,
            env=run_env,
            timeout=timeout,
            max_output_bytes=max_output_bytes,
        )


def run_program_get_output(
    full_cmd: list[str],
    cwd: str | None = None,
    env: dict | None = None,
    timeout: float | None = None,
    max_output_bytes: int = 32768,
) -> ProcessOutput:
    proc = subprocess.Popen(
        full_cmd,
        cwd=cwd,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    out_chunks_lock = Lock()
    out_chunks = {}

    def reader(name, stream):
        chunks = []
        total = 0
        while data := stream.read(8192):
            if total < max_output_bytes:
                keep = min(len(data), max_output_bytes - total)
                chunks.append(data[:keep])
                total += keep
        with out_chunks_lock:
            out_chunks[name] = b"".join(chunks)
        stream.close()

    threads: list[Thread] = []
    for name, stream in (("stdout", proc.stdout), ("stderr", proc.stderr)):
        t = Thread(target=reader, args=(name, stream))
        t.start()
        threads.append(t)

    try:
        proc.wait(timeout=timeout)
    except subprocess.TimeoutExpired:
        proc.kill()
        for t in threads:
            t.join()
        raise

    for t in threads:
        t.join()

    stdout = out_chunks["stdout"].decode(errors="replace")
    stderr = out_chunks["stderr"].decode(errors="replace")

    return ProcessOutput(returncode=proc.returncode, stdout=stdout, stderr=stderr)


def entrypoint():
    """
    Entrypoint to cap memory of an executable.
    """
    import os
    import resource
    import sys

    def set_limit(res, val):
        try:
            resource.setrlimit(res, (val, val))
        except ValueError:
            # Some limits may be too low to set as both soft+hard; try soft only then hard.
            r = resource.getrlimit(res)
            resource.setrlimit(res, (min(val, r[1]), min(val, r[1])))

    mem = int(sys.argv[1])
    # Cap virtual address space and data segment. macOS may enforce one better than the other.
    try:
        set_limit(resource.RLIMIT_AS, mem)
    except Exception:
        pass
    try:
        set_limit(resource.RLIMIT_DATA, mem)
    except Exception:
        pass

    # Optional CPU time cap
    cpu = os.environ.get("SB_CPU_SECONDS")
    if cpu:
        try:
            cpu = int(cpu)
            set_limit(resource.RLIMIT_CPU, cpu)
        except Exception:
            pass

    # Now exec the target program
    argv = sys.argv[2:]
    if not argv:
        print("shim: no program provided", file=sys.stderr)
        sys.exit(127)
    os.execvp(argv[0], argv)


if __name__ == "__main__":
    entrypoint()
