var grpc = require('grpc');
var protoLoader = require('@grpc/proto-loader');
var MongoClient = require('mongodb').MongoClient;
var url = "mongodb://localhost:27017/";


var packageDefinition = protoLoader.loadSync(
    `${__dirname}/backup.proto`,
    {keepCase: true,
     longs: String,
     enums: String,
     defaults: true,
     oneofs: true
    });

var backUpProto = grpc.loadPackageDefinition(packageDefinition).proto;

// Implements the CreateToken RPC method.
function onMessageIn(call, callback) {
  var request = call.request;

  // Process the business logic
  console.log("OnMessageIn printout: ",request)
  var emptyRes = {}

  MongoClient.connect(url, function(err, db) {
    if (err) throw err;
    var dbo = db.db("backup");
    // var myobj = { name: "Company Inc", address: "Highway 37" };
    dbo.collection("message").insertOne(request, function(err, res) {
      if (err) throw err;
      console.log("1 document inserted");
      db.close();
    });
  });

  callback(null, emptyRes);
}

// Implements the DeleteToken RPC method.
function isReady(call, callback) {
  var request = call.request;
  console.log("IsReady printout: ", request)

  // Process the business logic
  var statusRes = {
       ok: true
  }
  
  callback(null, statusRes);
}

// // Starts an RPC server that receives requests for the Payment service
function main() {
  console.log("Start-JS server with port","5005")
  var server = new grpc.Server();
  server.addService(backUpProto.Backup.service, {onMessageIn: onMessageIn, isReady: isReady});
  server.bind('0.0.0.0:5005', grpc.ServerCredentials.createInsecure());
  server.start();
}
main();
