from pathlib import Path
import pandas as pd
import numpy as np
from time import time
from exceptions import (
    FileExistException,
    FileNotExistException
)


BASE_DIR = Path(__file__).resolve().parent

class DTO:
    def __init__(
            self,
            save_dir=(BASE_DIR / 'IMDB_data_sets/filtered/'),
            read_dir=(BASE_DIR / 'IMDB_data_sets/'),
            default_chunksize: int=3_000_000
        ) -> None:
        """
            Parameters
            ----------
            save_dir : str, optional
                Folder location to save files (default is BASE_DIR / 'IMDB_data_sets/filtered/')
            get_dir : str, optional
                Folder location to get files (default is BASE_DIR / 'IMDB_data_sets/')
            default_chunksize : int, optional
                Default value to be used when chunksize is not given in methods that take
                chunksize parameters (default is 3_000_000)
        """

        self.save_dir = save_dir
        self.save_dir.mkdir(parents=True, exist_ok=True)
        self.read_dir = read_dir
        self.default_chunksize = default_chunksize

    def timing_decorator(func):
        def wrapper(*args, **kwargs):
            start_time = time()
            result = func(*args, **kwargs)
            print(f"Function {func.__name__} took {time() - start_time} seconds to run.")
            return result
        return wrapper

    def is_exist(self, file_dir: Path) -> None:
        """
            Parameters
            ----------
            file_dir : pathlib.Path
                File path

            Raises
            ------
            FileExistException
                If the file exists
        """
        
        if file_dir.is_file():
            raise FileExistException(f"file is exist: {file_dir}")

    def is_not_exist(self, file_dir: Path) -> None:
        """
            Parameters
            ----------
            file_dir : pathlib.Path
                File path

            Raises
            ------
            FileNotExistException
                If the file does not exist
        """

        if not file_dir.is_file():
            raise FileNotExistException(f"file is not exist: {file_dir}")

    def df2csv(
            self,
            df: pd.DataFrame,
            name: str,
            overwrite: bool=False,
            index: bool=False
        ) -> None:
        """
            Parameters
            ----------
            df : DataFrame 
                DataFrame object you want to save
            name : str
                The name you want to save the DataFrame object
            overwrite : bool, optional
                When True, overwrite if file exists (default is False)
            index : bool, optional
                Save index column or no (deafault is False)

            Raises
            ------
            FileExistException
                If the overwrite parameter is false and the file exists
        """

        if not overwrite:
            self.is_exist(self.save_dir / name)
        df.to_csv(self.save_dir / name, index=index)

    @timing_decorator
    def filter_tconst(
            self,
            name: str,
            title_types: list[str]=['movie', 'tvMovie'],
            chunksize: int=None
        ) -> list[str]:
        """
            Parameters
            ----------
            name : str
                Name of the basics file to be read
            title_type : list, optional
                'titleType' type of lines to be read from file (default is ['movie', 'tvMovie'])
            chunksize : int
                Chunk size for reading data (default is self.default_chunksize (default is 3_000_000)).

            Returns
            -------
            list
                A list of tconst

            Raises
            ------
            FileNotExistException
                If the file does not exist
        """

        self.is_not_exist(self.read_dir / name)
        if chunksize is None:
            chunksize = self.default_chunksize

        tconst_list = []

        with pd.read_csv(
                    self.read_dir / name,
                    sep=r'\t',
                    chunksize=chunksize,
                    engine='python',
                    usecols=['tconst', 'titleType'],
                    dtype={'tconst': str, 'titleType': str},
                    na_values='\\N') as reader:

            for i, r in enumerate(reader):
                tconst_list += list(r[r.titleType.isin(title_types)]['tconst'])
        return tconst_list

    def get_tconst(self, name: str) -> list[str]:
        """
            Parameters
            ----------
            name : str
                Name of the tconst file to be read

            Returns
            -------
            list
                A list of tconst

            Raises
            ------
            FileNotExistException
                If the file does not exist
        """
        
        self.is_not_exist(self.save_dir / name)
        return list(pd.read_csv(self.save_dir / name, usecols=['tconst'], dtype={'tconst': str})['tconst'])

    @timing_decorator
    def filter_principal(
            self,
            name: str,
            tconst_list: list[str],
            category_list: list[str]=['actress', 'actor', 'director', 'writer'],
            chunksize: int=None
        ) -> pd.DataFrame:
        """
            Parameters
            ----------
            name : str
                Name of the principals file to be read
            tconst_list : list
                List of tconst (It can be obtained by the get_tconst or read_tconst method).
            category : list
                List of categories of rows to be selected (default is ['actress', 'actor', 'director', 'writer']).
            chunksize : int
                Chunk size for reading data (default is self.default_chunksize (default is 3_000_000)).

            Returns
            -------
            DataFrame
                A DataFrame object with columns tconst, nconst, and category.

            Raises
            ------
            FileNotExistException
                If the file does not exist
        """

        self.is_not_exist(self.read_dir / name)
        if chunksize is None:
            chunksize = self.default_chunksize

        df = pd.DataFrame({
                'tconst': tconst_list,
                'nconst': np.empty((len(tconst_list), 0)).tolist(),
                'category': np.empty((len(tconst_list), 0)).tolist()
            })

        # index = pd.Index(tconst_list, name='tconst')
        # df = pd.DataFrame({
        #     'nconst': pd.Series(dtype='object', index=index),
        #     'category': pd.Series(dtype='object', index=index)
        # })

        cnt = 0

        with pd.read_csv(self.read_dir / name,
                        sep=r'\t',
                        chunksize=chunksize,
                        engine='python',
                        usecols=['tconst', 'nconst', 'category']) as reader:

            for i, r in enumerate(reader):
                r = r.query(f"(tconst in @tconst_list) and (category in @category_list)")
                r_group = r.groupby('tconst', as_index=0).agg({'nconst': lambda x: list(x), 'category': lambda x: list(x)})
                df = pd.concat([df, r_group]).groupby('tconst', as_index=0).agg(sum)

                # r_group.index.name = 'tconst'
                # df.update(r_group)
                del r_group

        print(cnt)
        return df

    @timing_decorator
    def filter_rating(
            self,
            name: str,
            tconst_list: list[str],
            chunksize: int=None
        ) -> pd.DataFrame:
        """
            Parameters
            ----------
            name : str
                Name of the ratings file to be read
            tconst_list : list
                List of tconst (It can be obtained by the get_tconst or read_tconst method).
            chunksize : int
                Chunk size for reading data (default is self.default_chunksize (default is 3_000_000)).

            Returns
            -------
            DataFrame
                A DataFrame object with columns tconst, and averageRating.

            Raises
            ------
            FileNotExistException
                If the file does not exist
        """

        self.is_not_exist(self.read_dir / name)
        if chunksize is None:
            chunksize = self.default_chunksize

        df = pd.DataFrame({'tconst': tconst_list})

        with pd.read_csv(
                self.read_dir / name,
                sep=r'\t',
                chunksize=chunksize,
                engine='python',
                usecols=['tconst', 'averageRating', 'numVotes'],
                dtype={'tconst': str, 'averageRating': np.float16, 'numVotes': int},
                na_values='\\N') as reader:

            for i, r in enumerate(reader):
                df = pd.concat([df, r.query("tconst in @tconst_list")]).groupby('tconst', as_index=0).first()
        return df

    @timing_decorator
    def filter_basic(
            self,
            name: str,
            tconst_list: list[str],
            chunksize: int=None
        ) -> pd.DataFrame:
        """
            Parameters
            ----------
            name : str
                Name of the basics file to be read
            tconst_list : list
                List of tconst (It can be obtained by the get_tconst or read_tconst method).
            chunksize : int
                Chunk size for reading data (default is self.default_chunksize (default is 3_000_000)).

            Returns
            -------
            DataFrame
                A DataFrame object with columns tconst, startYear and genres.

            Raises
            ------
            FileNotExistException
                If the file does not exist
        """

        self.is_not_exist(self.read_dir / name)
        if chunksize is None:
            chunksize = self.default_chunksize

        df = pd.DataFrame({'tconst': tconst_list})

        with pd.read_csv(self.read_dir / name,
                        sep=r'\t',
                        chunksize=chunksize,
                        engine='python',
                        usecols=['tconst', 'startYear', 'genres'],
                        dtype={'tconst': str, 'startYear': 'Int16', 'genres': str},
                        na_values='\\N') as reader:

            for i, r in enumerate(reader):
                df = pd.concat([df, r.query("tconst in @tconst_list")]).groupby('tconst', as_index=0).first()
        return df