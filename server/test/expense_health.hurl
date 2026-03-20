GET http://localhost:{{bff_port}}/api/v1/expenses/health
HTTP 200
[Asserts]
jsonpath "$.status" == "SERVING"
