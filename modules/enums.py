from enum import Enum


class AliasType(str, Enum):
    music = "music"
    character = "character"

    def __str__(self) -> str:
        return self.value
