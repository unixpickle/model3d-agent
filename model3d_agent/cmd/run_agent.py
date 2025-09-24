import argparse
import json
import os
from dataclasses import asdict
from uuid import uuid4

from openai import OpenAI

from model3d_agent.agent import compare_solutions, create_solution
from model3d_agent.chat import Chat, Solution
from model3d_agent.execution import CodeExecutor


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--prompt", type=str, required=True)
    parser.add_argument("--run_dir", type=str, required=True)
    parser.add_argument("--max_initial_attempts", type=int, default=5)
    args = parser.parse_args()

    run_dir: str = args.run_dir
    prompt: str = args.prompt

    os.makedirs(run_dir, exist_ok=True)

    client = OpenAI()
    executor = CodeExecutor()
    executor.warmup()
    try:
        if not (best_solution := find_current_best_solution(run_dir)):
            chat, solution = create_solution(
                client=client,
                executor=executor,
                prompt=prompt,
                max_attempts=args.max_initial_attempts,
            )
            id = save_solution(run_dir, chat, solution)
            if solution is None:
                print("failed to create initial model")
                return
            best_solution = solution
            mark_best_solution(run_dir, id)
        else:
            print("loaded current best solution.")

        while True:
            print("running refinement step")
            chat, solution = create_solution(
                client=client,
                executor=executor,
                prompt=prompt,
                previous=best_solution,
                max_attempts=args.max_initial_attempts,
            )
            id = save_solution(run_dir, chat, solution)
            print("output ID: " + id)
            if solution is None:
                print("refinement step failed")
            else:
                print("refinement step produced a new solution")
                comparison = compare_solutions(
                    client,
                    prompt,
                    a=best_solution,
                    b=solution,
                )
                if comparison == "b":
                    print("found better solution!")
                    mark_best_solution(run_dir, id)
                else:
                    print(
                        "did not find better solution, comparison yielded: "
                        + comparison
                    )
    finally:
        executor.cleanup()


def find_current_best_solution(run_dir: str) -> Solution | None:
    soln_name_file = os.path.join(run_dir, "best_solution")
    try:
        with open(soln_name_file, "r") as f:
            soln_name = f.read().strip().split("\n")[-1]
        return read_solution(os.path.join(run_dir, soln_name))
    except FileNotFoundError:
        return None


def read_solution(solution_dir: str) -> Solution:
    with open(os.path.join(solution_dir, "code.go"), "r") as f:
        code = f.read()
    with open(os.path.join(solution_dir, "rendering.jpg"), "rb") as f:
        rendering = f.read()
    return Solution(code=code, rendering=rendering)


def save_solution(run_dir: str, chat: Chat, solution: Solution | None) -> str:
    id = str(uuid4())
    os.makedirs(os.path.join(run_dir, id))
    if solution is not None:
        with open(os.path.join(run_dir, id, "rendering.jpg"), "wb") as f:
            f.write(solution.rendering)
        with open(os.path.join(run_dir, id, "code.go"), "w") as f:
            f.write(solution.code)
    with open(os.path.join(run_dir, id, "chat.json"), "w") as f:
        json.dump(asdict(chat), f)
    return id


def mark_best_solution(run_dir: str, id: str):
    try:
        with open(os.path.join(run_dir, "best_solution"), "r") as f:
            previous_best_solutions = f.read() + "\n"
    except FileNotFoundError:
        previous_best_solutions = ""

    with open(os.path.join(run_dir, "best_solution.tmp"), "w") as f:
        f.write(previous_best_solutions + id)
    os.rename(
        os.path.join(run_dir, "best_solution.tmp"),
        os.path.join(run_dir, "best_solution"),
    )
    soln = find_current_best_solution(run_dir)
    assert soln is not None


if __name__ == "__main__":
    main()
