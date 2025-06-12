import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "1s", target: 300 },
    { duration: "3s", target: 500 },
    { duration: "5s", target: 0 },
  ],
};

export default function () {
  const url = "http://127.0.0.1:8080/JCawU4I";

  const res = http.get(url);

  check(res, {
    "is status 200": (r) => r.status === 200,
  });

  sleep(0.1); // simulate user wait time between requests
}
