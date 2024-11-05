from typing import Any
import numpy as np
from sklearn.feature_extraction.text import CountVectorizer
from sklearn.metrics.pairwise import cosine_similarity
import polars as pl
from dataclasses import dataclass


@dataclass
class Filter:
    min_votes: int = None
    max_votes: int = None
    min_year: int = None
    max_year: int = None
    min_rating: float = None
    max_rating: float = None

    def __post_init__(self):
        if self.min_votes is not None and self.min_votes < 0:
            raise ValueError("min_votes should be greater than or equal to 0")
        if self.max_votes is not None and self.max_votes < 0:
            raise ValueError("max_votes should be greater than or equal to 0")
        if self.min_votes is not None and self.max_votes is not None and self.min_votes > self.max_votes:
            raise ValueError("min_votes should be less than or equal to max_votes")

        if self.min_year is not None and self.min_year < 0:
            raise ValueError("min_year should be greater than or equal to 0")
        if self.max_year is not None and self.max_year < 0:
            raise ValueError("max_year should be greater than or equal to 0")
        if self.min_year is not None and self.max_year is not None and self.min_year > self.max_year:
            raise ValueError("min_year should be less than or equal to max_year")

        if self.min_rating is not None and self.min_rating < 0:
            raise ValueError("min_rating should be greater than or equal to 0")
        if self.max_rating is not None and self.max_rating < 0:
            raise ValueError("max_rating should be greater than or equal to 0")
        if self.min_rating is not None and self.max_rating is not None and self.min_rating > self.max_rating:
            raise ValueError("min_rating should be less than or equal to max_rating")

@dataclass
class Weight:
    year: int = 100
    rating: int = 100
    genres: int = 100
    nconsts: int = 100

    def __post_init__(self):
        total_sum = 0
        total_count = 0
        for k, v in self.__dict__.items():
            if v < 0:
                raise ValueError(f'Weight for {k} must be greater than or equal to 0, got {v}')
            if v > 0:
                total_sum += v
                total_count += 1

        if total_sum < 100:
            raise ValueError(f'Total sum of weights must be at least 100, got {total_sum}')
        if total_count*100 != total_sum:
            raise ValueError(f'Total sum of weights must be {total_count*100}, got {total_sum}')

class Recommender:
    def __init__(
        self,
        filter_: Filter = Filter(),
        weight: Weight = Weight()
    ) -> None:
        self.filter = filter_
        self.weight = weight
        self.sql_where_clause = ''

        self.sql_where_clause = self.add_sql_where_clause(self.sql_where_clause, f"genres != ''")
        self.sql_where_clause = self.add_sql_where_clause(self.sql_where_clause, f"nconsts != ''")

        if filter_.min_votes:
            self.sql_where_clause = self.add_sql_where_clause(self.sql_where_clause, f'votes >= {filter_.min_votes}')
        if filter_.max_votes:
            self.sql_where_clause = self.add_sql_where_clause(self.sql_where_clause, f'votes <= {filter_.max_votes}')
        if filter_.min_year:
            self.sql_where_clause = self.add_sql_where_clause(self.sql_where_clause, f'year >= {filter_.min_year}')
        if filter_.max_year:
            self.sql_where_clause = self.add_sql_where_clause(self.sql_where_clause, f'year <= {filter_.max_year}')
        if filter_.min_rating:
            self.sql_where_clause = self.add_sql_where_clause(self.sql_where_clause, f'rating >= {filter_.min_rating}')
        if filter_.max_rating:
            self.sql_where_clause = self.add_sql_where_clause(self.sql_where_clause, f'rating <= {filter_.max_rating}')

    def add_sql_where_clause(self, old: str, new: str) -> None:
        return f'WHERE {new}' if old == '' else f'{old} AND {new}'

    def get_ordered_year_from_sql(self, conn, reference_year: int) -> pl.DataFrame:
        """
        Args
        ----
        conn: psycopg2 connection object
        reference_year: int - year to sort by closest

        Returns
        -------
        DataFrame:
        First sorted by closest year, then by number of votes (descending).
        | year_index (uint32) | tconst (str) |
        | ---                 | ---          |
        | 0                   | tt0000001    |
        | 1                   | tt0000002    |
        | 2                   | tt0000003    |
        | ...                 | ...          |
        """
        return pl.read_database(
            f"""
                SELECT tconst
                FROM imdb
                {self.sql_where_clause}
                ORDER BY ABS(year - {reference_year}), votes DESC
            """,
            conn, schema_overrides={'tconst': str}
        ).with_row_index('year_index')

    def get_ordered_rating_from_sql(self, conn, reference_rating: int) -> pl.DataFrame:
        """
        Args
        ----
        conn: psycopg2 connection object
        reference_rating: int - rating to sort by closest

        Returns
        -------
        DataFrame:
        First sorted by closest rating, then by number of votes (descending).
        | rating_index (uint32) | tconst (str) |
        | ---                   | ---          |
        | 0                     | tt0000001    |
        | 1                     | tt0000002    |
        | 2                     | tt0000003    |
        | ...                   | ...          |
        """
        return pl.read_database(
            f"""
                SELECT tconst
                FROM imdb
                {self.sql_where_clause}
                ORDER BY ABS(rating - {reference_rating}), votes DESC
            """,
            conn, schema_overrides={'tconst': str}
        ).with_row_index('rating_index')

    def get_ordered_genres_from_df(self, df: pl.DataFrame, reference_genres: str) -> pl.DataFrame:
        """
        Args
        ----
        df: DataFrame
        | tconst (str) | genres (str)   | votes (uint32) |
        | ---          | ---            | ---            |
        | tt0000001    | Drama, Romance | 123            |
        | tt0000002    | Comedy, Drama  | 456            |
        | tt0000003    | Action, Drama  | 789            |
        | ...          | ...            | ...            |
        reference_genres: str - genres to calculate cosine similarities

        Returns
        -------
        DataFrame:
        First sorted by cosine similarities genres (descending) and then by number of votes (descending).
        | genres_index (uint32) | tconst (str) |
        | ---                   | ---          |
        | 0                     | tt0000001    |
        | 1                     | tt0000002    |
        | 2                     | tt0000003    |
        | ...                   | ...          |
        """
        df = df.with_row_index('genres_index')

        genres_cv = CountVectorizer(dtype=np.uint8, token_pattern=r"(?u)[\w'-]+")
        genres_count_matrix = genres_cv.fit_transform(df['genres'])

        genres_sims = cosine_similarity(genres_cv.transform([reference_genres]), genres_count_matrix)[0]

        return pl.DataFrame(
            {
                'tconst': df['tconst'],
                'cosine_similarity': genres_sims,
                'votes': df['votes']
            }, schema={'tconst': str, 'cosine_similarity': pl.Float32, 'votes': pl.UInt32}
        ).\
            sort(['cosine_similarity', 'votes'], descending=True).\
                drop(['cosine_similarity', 'votes']).\
                    with_row_index('genres_index')

    def get_ordered_nconsts_from_df(self, df: pl.DataFrame, reference_nconsts: str) -> pl.DataFrame:
        """
        Args
        ----
        df: DataFrame
        | tconst (str) | nconsts (str)        | votes (uint32) |
        | ---          | ---                  | ---            |
        | tt0000001    | nm0000001, nm0000002 | 123            |
        | tt0000002    | nm0000001, nm0000003 | 456            |
        | tt0000003    | nm0000004, nm0000002 | 789            |
        | ...          | ...                  | ...            |
        reference_nconsts: str - nconsts to calculate cosine similarities

        Returns
        -------
        df: DataFrame
        First sorted by cosine similarities of nconsts (descending) and then by number of votes (descending).
        | nconsts_index (uint32) | tconst (str) |
        | ---                    | ---          |
        | 0                      | tt0000001    |
        | 1                      | tt0000002    |
        | 2                      | tt0000003    |
        | ...                    | ...          |
        """
        df = df.with_row_index('nconsts_index')

        nconsts_cv = CountVectorizer(dtype=np.uint8, token_pattern=r"(?u)[\w'-]+")
        nconsts_count_matrix = nconsts_cv.fit_transform(df['nconsts'])

        nconsts_sims = cosine_similarity(nconsts_cv.transform([reference_nconsts]), nconsts_count_matrix)[0]

        return  pl.DataFrame(
            {
                'tconst': df['tconst'],
                'cosine_similarity': nconsts_sims,
                'votes': df['votes']
            }, schema={'tconst': str, 'cosine_similarity': pl.Float32, 'votes': pl.UInt32}
        ).\
            sort(['cosine_similarity', 'votes'], descending=True).\
                drop(['cosine_similarity', 'votes']).\
                    with_row_index('nconsts_index')

    def get_main_df(self, conn) -> pl.DataFrame:
        """
        Args
        ----
        conn: psycopg2 connection object

        Returns
        -------
        DataFrame:
        | tconst (str) | genres (str)   | nconsts (str)        | votes (uint32) |
        | ---          | ---            | ---                  | ---            |
        | tt0000001    | Drama, Romance | nm0000001, nm0000002 | 123            |
        | tt0000002    | Comedy, Drama  | nm0000001, nm0000003 | 456            |
        | tt0000003    | Action, Drama  | nm0000004, nm0000002 | 789            |
        | ...          | ...            | ...                  | ...            |
        """
        return pl.read_database(
            f"""
                SELECT tconst, genres, nconsts, votes
                FROM imdb
                {self.sql_where_clause}
            """, conn, schema_overrides={'tconst': str, 'genres': str, 'nconsts': str, 'votes': pl.UInt32}
        )

    def get_row_by_tconst(self, conn, tconst: str) -> dict[str, Any]:
        """
        Args
        ----
        conn: psycopg2 connection object
        tconst: str - tconst to get row from database

        Returns
        -------
        dict: row from database
        {
            'tconst': str,
            'year': int,
            'genres': str,
            'nconsts': str,
            'rating': float,
            'votes': int
        }

        Raises
        ------
        ValueError: if tconst is not found in database
        """
        with conn.cursor() as cursor:
            cursor.execute(
                f"""
                    SELECT tconst, year, genres, nconsts, rating, votes
                    FROM imdb
                    WHERE tconst = '{tconst}'
                """
            )
            row = cursor.fetchone()
            if row is None:
                raise ValueError(f"tconst '{tconst}' not found")
            return {cursor.description[i][0]: value for i, value in enumerate(row)}

    def set_average(self, column_name: str, features: list[str], merged_df: pl.DataFrame) -> pl.DataFrame:
        """
        Args
        ----
        column_name: str - name of the column to store the average
        features: list[str] - list of features to calculate the average
        merged_df: DataFrame - merged DataFrame of all features

        Returns
        -------
        DataFrame: Same DataFrame with the argument column_name added to it with the average of all features
        """
        average = merged_df[f'{features[0]}_index'] * self.weight.__getattribute__(features[0])
        for feature in features[1:]:
            average += merged_df[f'{feature}_index'] * self.weight.__getattribute__(feature)

        return merged_df.with_columns(**{column_name: (average / (len(features) * 100))})

    def get_single_recommendation(self, conn, tconst: str, features: list[str]) -> pl.DataFrame:
        """
        Args
        ----
        conn: psycopg2 connection object
        tconst: str - tconst to get recommendations
        features: list[str] - list of features to calculate the average

        Returns
        -------
        DataFrame: DataFrame with the average of all features

        Raises
        ------
        ValueError: if no recommendations found
        """
        reference_row = self.get_row_by_tconst(conn, tconst)
        trained: dict[str, pl.DataFrame] = {}

        if 'year' in features:
            df = self.get_ordered_year_from_sql(conn, reference_year=reference_row['year'])
            if len(df) > 0:
                trained['year'] = df
        if 'rating' in features:
            df = self.get_ordered_rating_from_sql(conn, reference_rating=reference_row['rating'])
            if len(df) > 0:
                trained['rating'] = df
        if 'genres' in features or 'nconsts' in features:
            main_df = self.get_main_df(conn)
            if len(main_df) > 0:
                if 'genres' in features:
                    trained['genres'] = self.get_ordered_genres_from_df(
                                            pl.DataFrame(
                                                {
                                                    'tconst': main_df['tconst'],
                                                    'genres': main_df['genres'],
                                                    'votes': main_df['votes']
                                                }
                                            ), reference_genres=reference_row['genres']
                                        )
                if 'nconsts' in features:
                    trained['nconsts'] = self.get_ordered_nconsts_from_df(
                                            pl.DataFrame(
                                                {
                                                    'tconst': main_df['tconst'],
                                                    'nconsts': main_df['nconsts'],
                                                    'votes': main_df['votes']
                                                }
                                            ), reference_nconsts=reference_row['nconsts']
                                        )

        if len(trained) == 0:
            raise ValueError("No recommendations found, try changing the filter or weight")
        if len(features) > 1:
            merged = pl.concat(trained.values(), how='align')
            return self.set_average(
                    "average", features=features, merged_df=merged
                )
        else:
            trained_df = trained[features[0]]
            return trained_df.with_columns(
                    average=trained_df[f'{features[0]}_index']
                )

    def get_recommendations(self, conn, tconsts: list[str], n: int = 5) -> dict[str, list[str]]:
        """
        Args
        ----
        conn: psycopg2 connection object
        tconsts: list[str] - list of tconsts to get recommendations
        n: int - number of recommendations to get
        
        Returns
        -------
        list[dict[str, list[str]]]: list of dictionaries with tconst (ascending)
        as key and list of weights of columns as value (ascending)
        """
        self.sql_where_clause = self.add_sql_where_clause(
            self.sql_where_clause,
            f"tconst NOT IN ({', '.join(f"'{tconst}'" for tconst in tconsts)})"
        )

        features: list[str] = []
        if self.weight.year > 0:
            features.append('year')
        if self.weight.rating > 0:
            features.append('rating')
        if self.weight.genres > 0:
            features.append('genres')
        if self.weight.nconsts > 0:
            features.append('nconsts')

        if len(tconsts) == 1:
            merged_df = self.get_single_recommendation(conn, tconsts[0], features).sort('average')[:n]

            responses: dict[str, list[str]] = dict()
            for row in merged_df.rows(named=True):
                row.pop('average')
                t: str = row.pop('tconst')
                for f in features:
                    row[f] = row[f"{f}_index"] / self.weight.__getattribute__(f)
                    row.pop(f"{f}_index")
                weights: list[str] = [column for column, _ in sorted(row.items(), key=lambda item: item[1])]
                responses[t] = weights

            return responses
        else:
            trained_dfs: dict[str, pl.DataFrame] = {}
            for tconst in tconsts:
                df = self.get_single_recommendation(conn, tconst, features)
                trained_dfs[tconst] = pl.DataFrame({
                    'tconst': df['tconst'],
                    f"{tconst}_average": df['average']
                })

            merged_df: pl.DataFrame = pl.concat(trained_dfs.values(), how='align')

            all_average = merged_df[f"{tconsts[0]}_average"]
            for tconst in tconsts[1:]:
                all_average += merged_df[f"{tconst}_average"]
            merged_df = merged_df.with_columns(all_average=all_average / len(tconsts)).sort('all_average')[:n]

            responses: dict[str, list[str]] = dict()
            for row in merged_df.rows(named=True):
                row.pop('all_average')
                curretn_tconst: str = row.pop('tconst')
                for tconst in tconsts:
                    row[tconst] = row[f"{tconst}_average"]
                    row.pop(f"{tconst}_average")
                weights: list[str] = [column for column, _ in sorted(row.items(), key=lambda item: item[1])]
                responses[curretn_tconst] = weights

            return responses
