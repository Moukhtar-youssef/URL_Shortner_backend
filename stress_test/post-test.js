import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 100 }, // Gradual ramp-up
    { duration: "1m", target: 500 }, // Intermediate load
    { duration: "2m", target: 1000 }, // Peak load
    { duration: "30s", target: 0 }, // Ramp-down
  ],
  thresholds: {
    http_req_duration: ["p(95)<100"], // 95% of requests should be < 100ms
  },
};

export default function () {
  const url = "http://localhost:8080/create";
  const payload = JSON.stringify({
    long_url: "https://echo.labstack.com/docs/request",
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    "is status 201": (r) => r.status === 201,
  });

  sleep(0.1); // simulate user wait time between requests
}
