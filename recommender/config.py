import os

def get_postgres_dsn():
    user = os.getenv('POSTGRES_USER', None)
    if user is None:
        raise ValueError('POSTGRES_USER is not set')

    password = os.getenv('POSTGRES_PASSWORD', None)
    if password is None:
        raise ValueError('POSTGRES_PASSWORD is not set')

    host = os.getenv('POSTGRES_HOST', None)
    if host is None:
        raise ValueError('POSTGRES_HOST is not set')

    port = os.getenv('POSTGRES_PORT', None)
    if port is None:
        raise ValueError('POSTGRES_PORT is not set')
    try:
        port = int(port)
    except ValueError:
        raise ValueError('POSTGRES_PORT is not an integer')

    dbname = os.getenv('POSTGRES_DB', None)
    if dbname is None:
        raise ValueError('POSTGRES_DB is not set')

    return f'postgres://{user}:{password}@{host}:{port}/{dbname}?sslmode=disable'

def get_grpc_port():
    port = os.getenv('GRPC_PORT', None)
    if port is None:
        raise ValueError('GRPC_PORT is not set')
    try:
        port = int(port)
    except ValueError:
        raise ValueError('GRPC_PORT is not an integer')
    return port
