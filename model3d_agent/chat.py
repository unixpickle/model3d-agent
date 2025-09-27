import base64
import os
from dataclasses import dataclass
from functools import lru_cache

from .api_util import ChatMessage, ChatMessageContentImage

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
    def for_prompt_and_solution(
        cls, prompt: str, previous: Solution, critique: str
    ) -> "Chat":
        instructions = read_prompt_file("instructions.txt")
        docs = read_prompt_file("docs.txt")
        first_message = read_prompt_file("refinement.txt").format(
            prompt=prompt, docs=docs, previous=previous.code, critique=critique
        )
        return cls(
            instructions=instructions,
            messages=[
                {
                    "role": "user",
                    "content": [
                        {"type": "input_text", "text": first_message},
                        image_block(previous.rendering),
                    ],
                }
            ],
        )

    @classmethod
    def for_comparison(cls, prompt: str, a: bytes, b: bytes) -> "Chat":
        msg = read_prompt_file("compare.txt").format(prompt=prompt)
        return Chat(
            instructions="You are a helpful assistant which can identify and compare rendered 3D models.",
            messages=[
                {
                    "role": "user",
                    "content": [
                        {"type": "input_text", "text": msg},
                        image_block(a),
                        image_block(b),
                    ],
                },
            ],
        )

    @classmethod
    def for_critique(cls, prompt: str, img_data: bytes) -> "Chat":
        msg = read_prompt_file("critique.txt").format(prompt=prompt)
        return Chat(
            instructions="You are a helpful assistant which can identify, describe, and critique rendered 3D models.",
            messages=[
                {
                    "role": "user",
                    "content": [
                        {"type": "input_text", "text": msg},
                        image_block(img_data),
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

    def with_new_rendering(self, image: bytes) -> "Chat":
        return Chat(
            instructions=self.instructions,
            messages=self.messages
            + [
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "input_text",
                            "text": read_prompt_file("continue.txt"),
                        },
                        image_block(image),
                    ],
                }
            ],
        )


def image_block(data: bytes) -> ChatMessageContentImage:
    b64 = base64.b64encode(data).decode("ascii")
    return {
        "type": "input_image",
        "image_url": f"data:image/jpeg;base64,{b64}",
    }


@lru_cache()
def read_prompt_file(name: str) -> str:
    with open(os.path.join(PromptDir, name), "r") as f:
        return f.read()
