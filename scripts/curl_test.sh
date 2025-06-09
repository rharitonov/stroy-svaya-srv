curl -X POST http://localhost:8080/insertpdrline \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
	"pile_field_id": 1,
    "pile_number": "399",
    "start_date": "2025-04-05T14:30:00Z",
    "fact_pile_head": 10750,
	"recorded_by": "Иванов Иван Иванович"
  }'
  
  
 curl http://localhost:8080/getpdrlines?project_id=1
 curl http://localhost:8080/getpdrexcel?project_id=1
 curl http://localhost:8080/getpilestodriving?project_id=1
 curl http://localhost:8080/sendpdrlog?project_id=1
 curl http://localhost:8080/getuserfullname?tg_chat_id=?