from typing import Optional
from datetime import datetime
from pydantic import BaseModel, ConfigDict


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

    model_config = ConfigDict(from_attributes=True)


class UserPreferenceSchema(BaseModel):
    option: str
    value: str


class AliasBodySchema(BaseModel):
    im_id: str
    alias: str

    model_config = ConfigDict(from_attributes=True)


class AliasApprovalSchema(BaseModel):
    im_id: str


class AliasRejectionSchema(BaseModel):
    im_id: str
    reason: str


class PendingAliasSchema(BaseModel):
    id: int
    alias_type: str
    alias_type_id: int
    alias: str
    submitted_by: str
    submitted_at: datetime

    model_config = ConfigDict(from_attributes=True)
