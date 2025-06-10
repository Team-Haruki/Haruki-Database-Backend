from sqlalchemy import (
    Column,
    Integer,
    String,
    Boolean,
    DateTime,
    BigInteger,
    ForeignKey,
    UniqueConstraint
)
from sqlalchemy.orm import relationship

from .base import PjskBase


class UserBinding(PjskBase):
    __tablename__ = "user_bindings"
    __table_args__ = (
        UniqueConstraint("platform", "im_id", "server", name="uq_user_binding"),
    )

    id = Column(Integer, primary_key=True, autoincrement=True)
    platform = Column(String(20), nullable=False)
    im_id = Column(String(30), nullable=False, index=True)
    user_id = Column(String(30), nullable=False)
    server = Column(String(2), nullable=False)
    visible = Column(Boolean, default=True)

    default_refs = relationship(
        "UserDefaultBinding",
        back_populates="binding",
        cascade="all, delete",
        lazy="joined",
    )


class UserDefaultBinding(PjskBase):
    __tablename__ = "user_default_bindings"
    __table_args__ = (
        UniqueConstraint("im_id", "platform", "server", name="uq_user_default_binding"),
    )

    id = Column(Integer, primary_key=True, autoincrement=True)
    im_id = Column(String(30), nullable=False)
    platform = Column(String(20), nullable=False)
    server = Column(String(7), nullable=False)
    binding_id = Column(Integer, ForeignKey("user_bindings.id", ondelete="CASCADE"), nullable=False)

    binding = relationship("UserBinding", back_populates="default_refs")


class UserPreference(PjskBase):
    __tablename__ = "user_preferences"
    im_id = Column(String(30), primary_key=True)
    platform = Column(String(20), primary_key=True)
    option = Column(String(50), primary_key=True)
    value = Column(String(50), nullable=False)


class Alias(PjskBase):
    __tablename__ = "aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    alias_type = Column(String(20), nullable=False)  # e.g., "music", "character"
    alias_type_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)


class PendingAlias(PjskBase):
    __tablename__ = "pending_aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    alias_type = Column(String(20), nullable=False)
    alias_type_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)
    submitted_by = Column(String(100), nullable=False)
    submitted_at = Column(DateTime, nullable=False)


class RejectedAlias(PjskBase):
    __tablename__ = "rejected_aliases"
    id = Column(BigInteger, primary_key=True)
    alias_type = Column(String(20), nullable=False)
    alias_type_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)
    reviewed_by = Column(String(100), nullable=False)
    reason = Column(String(255), nullable=False)
    reviewed_at = Column(DateTime, nullable=False)


class GroupAlias(PjskBase):
    __tablename__ = "group_aliases"
    group_id = Column(String(50), nullable=False, primary_key=True)
    alias_type = Column(String(20), nullable=False, primary_key=True)
    alias_type_id = Column(Integer, nullable=False, primary_key=True)
    alias = Column(String(100), nullable=False, primary_key=True)


class AliasAdmin(PjskBase):
    __tablename__ = "alias_admins"
    im_id = Column(String(100), primary_key=True)
    name = Column(String(100), nullable=False)
