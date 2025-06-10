from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, ConfigDict, Field

from modules.enums import DefaultBindingServer


class BaseSchema(BaseModel):
    message: Optional[str] = "success"
    status: Optional[int] = 200


class EditBindingSchema(BaseModel):
    server: Optional[DefaultBindingServer] = DefaultBindingServer.jp
    user_id: Optional[str] = None
    binding_id: Optional[int] = None
    visible: Optional[bool] = True


class BindingSchema(BaseModel):
    id: int
    platform: str
    im_id: str
    server: str
    user_id: str
    visible: bool

    model_config = ConfigDict(from_attributes=True)


class AliasSchema(BaseModel):
    alias: str
    im_id: Optional[str] = None


class UserPreferenceSchema(BaseModel):
    option: Optional[str] = None
    value: Optional[str] = None


class PendingAliasSchema(BaseModel):
    id: int = Field(..., description="Pending alias ID")
    alias_type: str = Field(..., description="Alias type")
    alias_type_id: int = Field(..., description="ID of the target entity for the alias")
    alias: str = Field(..., description="Alias value")
    submitted_by: str = Field(..., description="User ID of the submitter")
    submitted_at: datetime = Field(..., description="Submission time")

    model_config = ConfigDict(from_attributes=True)


class AliasApprovalSchema(BaseModel):
    im_id: str


class AliasRejectionSchema(BaseModel):
    im_id: str
    reason: str


class BindingResponse(BaseSchema):
    bindings: Optional[List[BindingSchema]] = None
    binding: Optional[BindingSchema] = None


class AddBindingSuccessResponse(BaseSchema):
    bind_id: Optional[int] = None


class UserPreferenceResponse(BaseSchema):
    options: Optional[List[UserPreferenceSchema]] = None
    option: Optional[UserPreferenceSchema] = None


class AliasToObjectIdResponse(BaseSchema):
    match_ids: Optional[List[int]] = None


class AllAliasesResponse(BaseSchema):
    aliases: Optional[List[str]] = None


class PendingAliasListResponse(BaseSchema):
    rows: int
    results: List[PendingAliasSchema]
