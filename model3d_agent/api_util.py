import time
from typing import Any, Literal, NotRequired, TypedDict

from openai import OpenAI, RateLimitError


class CompletionError(Exception):
    pass


class ChatMessageContentText(TypedDict):
    type: Literal["input_text", "output_text"]
    text: str


class ChatMessageContentImage(TypedDict):
    type: Literal["input_image"]
    image_url: str
    detail: NotRequired[Literal["low", "high", "auto"]]


ChatMessageContent = ChatMessageContentText | ChatMessageContentImage


class ChatMessage(TypedDict):
    role: Literal["user", "assistant"]
    content: str | list[ChatMessageContent]


def completion(client: OpenAI, instructions: str, input: Any) -> str:
    limit_hits = 0
    while True:
        try:
            response = client.responses.create(
                model="gpt-4.1",
                instructions=instructions,
                input=input,
            )
        except RateLimitError as exc:
            limit_hits += 1
            delay = min(30, 2**limit_hits)
            print(f"hit rate limit: {exc}")
            print(f"waiting {delay} seconds")
            time.sleep(delay)
            continue
        except KeyboardInterrupt:
            raise
        except Exception as exc:
            raise CompletionError("API call failed") from exc
        if err := response.error:
            raise CompletionError(f"error: {err}")
        return response.output_text
