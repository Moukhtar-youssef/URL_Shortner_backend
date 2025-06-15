import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "5s", target: 50 }, // ramp up to 50 VUs
    { duration: "10s", target: 50 }, // stay at 50 VUs
    { duration: "5s", target: 0 }, // ramp down to 0
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
