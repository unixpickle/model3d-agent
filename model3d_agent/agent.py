import re
from typing import Literal

from openai import OpenAI

from .api_util import completion
from .chat import Chat, Solution
from .execution import CodeExecutor, ExecutionError


def create_solution(
    client: OpenAI,
    executor: CodeExecutor,
    prompt: str,
    previous: Solution | None = None,
    max_attempts: int = 5,
    verbose: bool = False,
    max_refinements: int = 2,
) -> tuple[Chat, Solution | None]:
    if previous:
        critique_chat = Chat.for_critique(prompt=prompt, img_data=previous.rendering)
        critique = completion(
            client, critique_chat.instructions, critique_chat.messages
        )
        if verbose:
            print(f" * created critique of length {len(critique)}")
    chat = (
        Chat.for_prompt(prompt)
        if previous is None
        else Chat.for_prompt_and_solution(prompt, previous=previous, critique=critique)
    )
    solution = None
    num_refinements = 0
    for i in range(max_attempts):
        response = completion(client, chat.instructions, chat.messages)
        chat = chat.with_response(response)
        matches = re.findall(r"```(?:[a-zA-Z]*)\n(.*?)```", response, re.DOTALL)
        if not len(matches):
            if verbose:
                print(" * no code blocks in response")
            chat = chat.with_error("No code blocks were found in the response.")
            continue
        code = matches[-1]
        try:
            rendering = executor.execute_code(code)
        except ExecutionError as err:
            if verbose:
                print(f" * error executing code: {err.message.splitlines()[0]}")
            chat = chat.with_error(err.message)
            continue
        solution = Solution(code=code, rendering=rendering)
        num_refinements += 1
        if verbose:
            print(" * got solution with rendering")
        if num_refinements > max_refinements:
            break
        if i + 1 < max_attempts:
            chat = chat.with_new_rendering(rendering)
    return chat, solution


def compare_solutions(
    client: OpenAI,
    prompt: str,
    a: Solution,
    b: Solution,
) -> Literal["a", "b", "tie", "unknown"]:
    chat = Chat.for_comparison(prompt, a.rendering, b.rendering)
    response = completion(client, chat.instructions, chat.messages)
    if response.endswith("Answer: first"):
        return "a"
    elif response.endswith("Answer: second"):
        return "b"
    elif response.endswith("Answer: same"):
        return "tie"
    else:
        return "unknown"
