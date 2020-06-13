service_dir="$(cd ../.. && pwd)"

for casefile in "${service_dir}"/loadtesting/DelCities/*.json; do
  ghz --insecure --proto="${service_dir}"/proto/cityservice.proto --call proto.CityService.DelCities 127.0.0.1:50051 -c 100 -n 2000 -D "${casefile}"
done