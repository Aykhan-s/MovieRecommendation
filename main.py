from IMDB_DTO import DTO
from time import time
import pandas as pd


if __name__ == '__main__':
    start_time = time()
    dto = DTO()

    tconst_list = dto.filter_tconst(name='title.basics.tsv')
    dto.df2csv(df=pd.DataFrame({'tconst': tconst_list}), name='tconst.csv')

    tconst_list = dto.get_tconst('tconst.csv')

    df = dto.filter_basic(name='title.basics.tsv', tconst_list=tconst_list)
    dto.df2csv(df=df, name='basics.csv')
    del df
    df = dto.filter_principal(name='title.principals.tsv', tconst_list=tconst_list)
    dto.df2csv(df=df, name='principals_comma.csv', overwrite=True)
    del df
    df = dto.filter_rating(name='title.ratings.tsv', tconst_list=tconst_list)
    dto.df2csv(df=df, name='ratings.csv', overwrite=1)
    del df