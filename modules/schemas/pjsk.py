from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, ConfigDict, Field


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


class AliasToObjectIdSchema(BaseModel):
    message: Optional[str] = "success"
    code: Optional[int] = 200
    match_ids: Optional[List[int]] = None


class AllAliasesSchema(BaseModel):
    message: Optional[str] = "success"
    code: Optional[int] = 200
    aliases: Optional[List[str]] = None


class AliasSchema(BaseModel):
    alias: str
    im_id: Optional[str] = None


class PendingAliasEntry(BaseModel):
    id: int = Field(..., description="Pending alias ID")
    alias_type: str = Field(..., description="Alias type")
    alias_type_id: int = Field(..., description="ID of the target entity for the alias")
    alias: str = Field(..., description="Alias value")
    submitted_by: str = Field(..., description="User ID of the submitter")
    submitted_at: datetime = Field(..., description="Submission time")

    model_config = ConfigDict(from_attributes=True)


class PendingAliasList(BaseModel):
    rows: int
    message: Optional[str] = "success"
    code: Optional[int] = 200
    results: List[PendingAliasEntry]


class AliasApprovalSchema(BaseModel):
    im_id: str


class AliasRejectionSchema(BaseModel):
    im_id: str
    reason: str
