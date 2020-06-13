service_dir="$(cd ../.. && pwd)"

for casefile in "${service_dir}"/loadtesting/AddCities/*.json; do
  ghz --insecure --proto="${service_dir}"/proto/cityservice.proto -O html --call proto.CityService.AddCities 127.0.0.1:50051 -c 100 -n 2000 -D "${casefile}"
done