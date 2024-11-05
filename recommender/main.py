from sys import path
path.append('./proto')

from concurrent import futures
from time import sleep
import threading
from recommend import Recommender, Weight, Filter
from config import get_postgres_dsn, get_grpc_port

import psycopg2

from proto import recommender_pb2, recommender_pb2_grpc
import grpc
from grpc_reflection.v1alpha import reflection
from grpc_health.v1 import health
from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc

postgres_dsn = get_postgres_dsn()

class RecommenderServicer(recommender_pb2_grpc.RecommenderServicer):
    def GetRecommendations(self, request: recommender_pb2.Request, context):
        try:
            recommender = Recommender(
                filter_=Filter(
                    min_votes=request.filter.min_votes if request.filter.HasField('min_votes_oneof') else None,
                    max_votes=request.filter.max_votes if request.filter.HasField('max_votes_oneof') else None,
                    min_year=request.filter.min_year if request.filter.HasField('min_year_oneof') else None,
                    max_year=request.filter.max_year if request.filter.HasField('max_year_oneof') else None,
                    min_rating=request.filter.min_rating if request.filter.HasField('min_rating_oneof') else None,
                    max_rating=request.filter.max_rating if request.filter.HasField('max_rating_oneof') else None
                ),
                weight=Weight(
                    year=request.weight.year,
                    rating=request.weight.rating,
                    genres=request.weight.genres,
                    nconsts=request.weight.nconsts
                )
            )
        except ValueError as e:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details(str(e))
            return recommender_pb2.Response()
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return recommender_pb2.Response()

        with psycopg2.connect(dsn=postgres_dsn) as conn:
            try:
                data = recommender.get_recommendations(conn, request.tconsts, request.n)
            except ValueError as e:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details(str(e))
                return recommender_pb2.Response()
            except Exception as e:
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details(str(e))
                return recommender_pb2.Response()

            movies = []
            for k, v in data.items():
                movies.append(
                    recommender_pb2.RecommendedMovie(
                        tconst=k,
                        weights=v
                    )
                )

        return recommender_pb2.Response(movies=movies)

def _toggle_health(health_servicer: health.HealthServicer, service: str):
    next_status = health_pb2.HealthCheckResponse.SERVING
    while True:
        if next_status == health_pb2.HealthCheckResponse.SERVING:
            next_status = health_pb2.HealthCheckResponse.NOT_SERVING
        else:
            next_status = health_pb2.HealthCheckResponse.SERVING

        health_servicer.set(service, next_status)
        sleep(5)

def _configure_health_server(server: grpc.Server):
    health_servicer = health.HealthServicer(
        experimental_non_blocking=True,
        experimental_thread_pool=futures.ThreadPoolExecutor(max_workers=10),
    )
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)

    toggle_health_status_thread = threading.Thread(
        target=_toggle_health,
        args=(health_servicer, "recommender.Recommender"),
        daemon=True,
    )
    toggle_health_status_thread.start()

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=100))
    recommender_pb2_grpc.add_RecommenderServicer_to_server(RecommenderServicer(), server)
    SERVICE_NAMES = (
        recommender_pb2.DESCRIPTOR.services_by_name["Recommender"].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(SERVICE_NAMES, server)
    server.add_insecure_port(f'[::]:{get_grpc_port()}')
    _configure_health_server(server)
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    try:
        serve()
    except KeyboardInterrupt:
        print("Shutting down server")
