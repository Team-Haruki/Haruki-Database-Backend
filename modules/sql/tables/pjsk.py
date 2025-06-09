from sqlalchemy import (
    Column,
    Integer,
    BigInteger,
    String,
    Boolean,
    DateTime,
    ForeignKey,
)
from sqlalchemy.orm import relationship

from .base import Base


class UserBinding(Base):
    __tablename__ = "user_bindings"
    id = Column(Integer, primary_key=True, autoincrement=True)
    platform = Column(String, primary_key=True)
    im_id = Column(String, nullable=False, index=True)
    user_id = Column(String, nullable=False)
    server = Column(String, nullable=False)
    visible = Column(Boolean, default=True)
    default_refs = relationship(
        "UserDefaultBinding",
        back_populates="binding",
        cascade="all, delete",
        lazy="joined",
    )


class UserDefaultBinding(Base):
    __tablename__ = "user_default_bindings"
    im_id = Column(String, primary_key=True)
    platform = Column(String, primary_key=True)
    server = Column(String, primary_key=True)  # 'jp', 'cn', ..., or 'default'
    bind_id = Column(Integer, ForeignKey("user_bindings.id", ondelete="CASCADE"), nullable=False)
    binding = relationship("UserBinding", back_populates="default_refs")


class UserPreference(Base):
    __tablename__ = "user_preferences"
    im_id = Column(String, primary_key=True)
    option = Column(String(50), primary_key=True)
    value = Column(String(50), nullable=False)


class Alias(Base):
    __tablename__ = "aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    alias_type = Column(String(20), nullable=False)  # e.g., "music", "character"
    alias_type_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)


class PendingAlias(Base):
    __tablename__ = "pending_aliases"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    alias_type = Column(String(20), nullable=False)
    alias_type_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)
    submitted_by = Column(String(100), nullable=False)
    submitted_at = Column(DateTime, nullable=False)


class RejectedAlias(Base):
    __tablename__ = "rejected_aliases"
    id = Column(BigInteger, primary_key=True)
    alias_type = Column(String(20), nullable=False)
    alias_type_id = Column(Integer, nullable=False)
    alias = Column(String(100), nullable=False)
    reviewed_by = Column(String(100), nullable=False)
    reason = Column(String(255), nullable=False)
    reviewed_at = Column(DateTime, nullable=False)


class GroupAlias(Base):
    __tablename__ = "group_aliases"
    group_id = Column(String(50), nullable=False, primary_key=True)
    alias_type = Column(String(20), nullable=False, primary_key=True)
    alias_type_id = Column(Integer, nullable=False, primary_key=True)
    alias = Column(String(100), nullable=False, primary_key=True)


class AliasAdmin(Base):
    __tablename__ = "alias_admins"
    im_id = Column(String(100), primary_key=True)
    name = Column(String(100), nullable=False)
