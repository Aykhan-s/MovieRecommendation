### Movier: Get Movie Recommendations Based on IMDb Data

Movier recommends movies based on the IMDb dataset by comparing them with movies you provide.

:warning: This project is not production-ready.

## Data Source
This project uses non-commercial IMDb datasets, available at: [IMDb Datasets](https://datasets.imdbws.com/)

## How to Run

1. Rename the environment file:
   Rename `/config/postgres/.env.example` to `/config/postgres/.env`:
   ```sh
   mv ./config/postgres/.env.example ./config/postgres/.env
   ```

2. Run with Docker Compose:
   Use Docker Compose to start the application:
   ```sh
   docker compose up
   ```

   *Downloading and extracting the datasets may take 5-10 minutes, depending on your network speed.*

3. Access the application:
   Once the setup is complete, go to [http://localhost:8080](http://localhost:8080) to access Movier.
