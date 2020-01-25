function InitCluster(){
    console.log(host);
    tosend = {Req: 0};
        //send the request
    $.post(host+"/initcluster",JSON.stringify(tosend)).done(function(data) {
        //update the peer list...
        console.log("OK for search");
    });


}