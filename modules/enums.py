from enum import Enum


class AliasType(str, Enum):
    music = "music"
    character = "character"

    def __str__(self) -> str:
        return self.value


class BindingServer(str, Enum):
    jp = "jp"
    en = "en"
    tw = "tw"
    kr = "kr"
    cn = "cn"

    def __str__(self) -> str:
        return self.value


class DefaultBindingServer(str, Enum):
    jp = "jp"
    en = "en"
    tw = "tw"
    kr = "kr"
    cn = "cn"

    def __str__(self) -> str:
        return self.value

    default = "default"
