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
) -> tuple[Chat, Solution | None]:
    chat = (
        Chat.for_prompt(prompt)
        if previous is None
        else Chat.for_prompt_and_solution(prompt, previous=previous)
    )
    for _ in range(max_attempts):
        response = completion(client, chat.instructions, chat.messages)
        chat = chat.with_response(response)
        matches = re.findall(r"```(?:[a-zA-Z]*)\n(.*?)```", response, re.DOTALL)
        if not len(matches):
            chat = chat.with_error("No code blocks were found in the response.")
            continue
        code = matches[-1]
        try:
            rendering = executor.execute_code(code)
        except ExecutionError as err:
            chat = chat.with_error(err.message)
            continue
        return chat, Solution(code=code, rendering=rendering)
    return chat, None


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
