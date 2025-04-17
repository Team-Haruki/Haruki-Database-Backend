from typing import Optional
from pydantic import BaseModel
from modules.sql.tables.pjsk import UserBinding


class AddBindingSchema(BaseModel):
    user_id: str
    visible: Optional[bool] = True


class SetDefaultBindingSchema(BaseModel):
    bind_id: int


class UpdateBindingVisibilitySchema(BaseModel):
    visible: bool


class BindingResult(BaseModel):
    id: int
    server: str
    user_id: str
    visible: bool

    @classmethod
    def from_orm(cls, obj: UserBinding):
        return cls(
            id=obj.id,
            server=obj.server,
            user_id=obj.user_id,
            visible=obj.visible
        )


class DefaultBindingResult(BaseModel):
    id: int
    server: str
    user_id: str
    visible: bool

    @classmethod
    def from_orm(cls, obj: UserBinding):
        return cls(
            id=obj.id,
            server=obj.server,
            user_id=obj.user_id,
            visible=obj.visible
        )


class UserPreferenceSchema(BaseModel):
    option: str
    value: str


class AliasBodySchema(BaseModel):
    alias: str
    group_id: Optional[str] = None
