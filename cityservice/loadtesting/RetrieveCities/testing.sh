service_dir="$(cd ../.. && pwd)"

for casefile in "${service_dir}"/loadtesting/RetrieveCities/*.json; do
  ghz --insecure --proto="${service_dir}"/proto/cityservice.proto -O html --call proto.CityService.RetrieveCities 127.0.0.1:50051 -c 100 -n 2000 -D "${casefile}"
done
