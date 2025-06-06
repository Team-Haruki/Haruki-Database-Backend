from datetime import datetime
from typing import Optional
from pydantic import BaseModel
from modules.sql.tables.pjsk import UserBinding, PendingAlias


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
        return cls(id=obj.id, server=obj.server, user_id=obj.user_id, visible=obj.visible)


class UserPreferenceSchema(BaseModel):
    option: str
    value: str


class AliasBodySchema(BaseModel):
    im_id: str
    alias: str


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

    @classmethod
    def from_orm(cls, obj: PendingAlias):
        return cls(
            id=obj.id,
            alias_type=obj.alias_type,
            alias_type_id=obj.alias_type_id,
            alias=obj.alias,
            submitted_by=obj.submitted_by,
            submitted_at=obj.submitted_at,
        )
