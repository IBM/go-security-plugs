import http from 'k6/http';
import { sleep, check } from 'k6';

export default function () {
  const res = http.get('http://127.0.0.1:8888');
  check(res, {
    'is status 200': (r) => r.status === 200,
  });
  console.log("STATUS......", res.status)
  sleep(1);
}
