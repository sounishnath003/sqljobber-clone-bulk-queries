import http from "k6/http";
import { group, sleep, check } from "k6";

export const options = {
  // A number specifying the number of VUs to run concurrently.
  vus: 10,
  // A string specifying the total duration of the test run.
  duration: "8s",
};

// The function that defines VU logic.
//
// See https://grafana.com/docs/k6/latest/examples/get-started-with-k6/ to learn more
// about authoring k6 scripts.
//

const API_URL = "http://localhost:3000";

export default function () {
  group("/healthy", function () {
    const response = http.get(`${API_URL}/healthy`);
    check(response, {
      "response code was 200?": (res) => res.status === 200,
      "response json as status 200?": (res) =>
        JSON.parse(res.body)["status"] == 200,
    });
    sleep(1);
  });

  group("/api/v1/tasks", function () {
    const response = http.get(`${API_URL}/api/v1/tasks`);
    check(response, {
      "/api/v1/tasks response code was 200?": (res) => res.status === 200,
      "/api/v1/tasks response json as status 200?": (res) =>
        JSON.parse(res.body)["status"] == 200,
    });
  });
  sleep(1);
}
