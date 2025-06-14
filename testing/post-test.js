import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "1s", target: 300 },
    { duration: "3s", target: 500 },
    { duration: "5s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<100"], // 95% of requests should be < 100ms
  },
};

export default function () {
  const url = "http://localhost:8081/create";
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
    "is status 200": (r) => r.status === 200,
  });

  sleep(0.1); // simulate user wait time between requests
}
