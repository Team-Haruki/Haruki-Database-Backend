from typing import Optional
from pydantic import BaseModel


class AddBindingSchema(BaseModel):
    user_id: str
    visible: Optional[bool] = True


class SetDefaultSchema(BaseModel):
    bind_id: int


class UpdateVisibilitySchema(BaseModel):
    visible: bool


class BindingOutSchema(BaseModel):
    id: int
    server: str
    user_id: str
    visible: bool


class DefaultOutSchema(BaseModel):
    id: int
    server: str
    user_id: str
    visible: bool
