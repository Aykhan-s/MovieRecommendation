from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Filter(_message.Message):
    __slots__ = ("min_votes", "max_votes", "min_year", "max_year", "min_rating", "max_rating")
    MIN_VOTES_FIELD_NUMBER: _ClassVar[int]
    MAX_VOTES_FIELD_NUMBER: _ClassVar[int]
    MIN_YEAR_FIELD_NUMBER: _ClassVar[int]
    MAX_YEAR_FIELD_NUMBER: _ClassVar[int]
    MIN_RATING_FIELD_NUMBER: _ClassVar[int]
    MAX_RATING_FIELD_NUMBER: _ClassVar[int]
    min_votes: int
    max_votes: int
    min_year: int
    max_year: int
    min_rating: float
    max_rating: float
    def __init__(self, min_votes: _Optional[int] = ..., max_votes: _Optional[int] = ..., min_year: _Optional[int] = ..., max_year: _Optional[int] = ..., min_rating: _Optional[float] = ..., max_rating: _Optional[float] = ...) -> None: ...

class Weight(_message.Message):
    __slots__ = ("year", "rating", "genres", "nconsts")
    YEAR_FIELD_NUMBER: _ClassVar[int]
    RATING_FIELD_NUMBER: _ClassVar[int]
    GENRES_FIELD_NUMBER: _ClassVar[int]
    NCONSTS_FIELD_NUMBER: _ClassVar[int]
    year: int
    rating: int
    genres: int
    nconsts: int
    def __init__(self, year: _Optional[int] = ..., rating: _Optional[int] = ..., genres: _Optional[int] = ..., nconsts: _Optional[int] = ...) -> None: ...

class Request(_message.Message):
    __slots__ = ("tconsts", "n", "filter", "weight")
    TCONSTS_FIELD_NUMBER: _ClassVar[int]
    N_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    WEIGHT_FIELD_NUMBER: _ClassVar[int]
    tconsts: _containers.RepeatedScalarFieldContainer[str]
    n: int
    filter: Filter
    weight: Weight
    def __init__(self, tconsts: _Optional[_Iterable[str]] = ..., n: _Optional[int] = ..., filter: _Optional[_Union[Filter, _Mapping]] = ..., weight: _Optional[_Union[Weight, _Mapping]] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ("movies",)
    MOVIES_FIELD_NUMBER: _ClassVar[int]
    movies: _containers.RepeatedCompositeFieldContainer[RecommendedMovie]
    def __init__(self, movies: _Optional[_Iterable[_Union[RecommendedMovie, _Mapping]]] = ...) -> None: ...

class RecommendedMovie(_message.Message):
    __slots__ = ("tconst", "weights")
    TCONST_FIELD_NUMBER: _ClassVar[int]
    WEIGHTS_FIELD_NUMBER: _ClassVar[int]
    tconst: str
    weights: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, tconst: _Optional[str] = ..., weights: _Optional[_Iterable[str]] = ...) -> None: ...
