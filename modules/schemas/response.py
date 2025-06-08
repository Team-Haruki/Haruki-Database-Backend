from pydantic import BaseModel
from typing import Optional


class APIResponse(BaseModel):
    status: Optional[int] = 200
    message: Optional[str] = "success"
