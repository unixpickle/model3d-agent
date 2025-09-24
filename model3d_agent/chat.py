import base64
import os
from dataclasses import dataclass
from functools import lru_cache

from .api_util import ChatMessage
from .execution import ExecutionError

PromptDir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "prompts")


@dataclass(frozen=True)
class Solution:
    code: str
    rendering: bytes


@dataclass(frozen=True)
class Chat:
    instructions: str
    messages: list[ChatMessage]

    @classmethod
    def for_prompt(cls, prompt: str) -> "Chat":
        instructions = read_prompt_file("instructions.txt")
        docs = read_prompt_file("docs.txt")
        first_message = read_prompt_file("first_message.txt").format(
            prompt=prompt, docs=docs
        )
        return cls(
            instructions=instructions,
            messages=[{"role": "user", "content": first_message}],
        )

    @classmethod
    def for_prompt_and_solution(cls, prompt: str, previous: Solution) -> "Chat":
        instructions = read_prompt_file("instructions.txt")
        docs = read_prompt_file("docs.txt")
        first_message = read_prompt_file("refinement.txt").format(
            prompt=prompt, docs=docs, previous=previous.code
        )
        render_b64 = base64.b64encode(previous.rendering).decode("ascii")
        return cls(
            instructions=instructions,
            messages=[
                {
                    "role": "user",
                    "content": [
                        {"type": "input_text", "text": first_message},
                        {
                            "type": "input_image",
                            "image_url": f"data:image/jpeg;base64,{render_b64}",
                            "detail": "low",
                        },
                    ],
                }
            ],
        )

    @classmethod
    def for_comparison(cls, prompt: str, a: bytes, b: bytes) -> "Chat":
        msg = read_prompt_file("compare.txt").format(prompt=prompt)
        b64_a = base64.b64encode(a).decode("ascii")
        b64_b = base64.b64encode(b).decode("ascii")
        return Chat(
            instructions="You are a helpful assistant which can identify and compare rendered 3D models.",
            messages=[
                {
                    "role": "user",
                    "content": [
                        {"type": "input_text", "text": msg},
                        {
                            "type": "input_image",
                            "image_url": f"data:image/jpeg;base64,{b64_a}",
                        },
                        {
                            "type": "input_image",
                            "image_url": f"data:image/jpeg;base64,{b64_b}",
                        },
                    ],
                },
            ],
        )

    def with_response(self, response: str) -> "Chat":
        return Chat(
            instructions=self.instructions,
            messages=self.messages + [{"role": "assistant", "content": response}],
        )

    def with_error(self, err: str) -> "Chat":
        return Chat(
            instructions=self.instructions,
            messages=self.messages
            + [
                {
                    "role": "user",
                    "content": "The above code failed to run with the following error. Please try again.\n\n"
                    + err,
                }
            ],
        )


@lru_cache()
def read_prompt_file(name: str) -> str:
    with open(os.path.join(PromptDir, name), "r") as f:
        return f.read()
