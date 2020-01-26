function InitCluster(){
    console.log(host);
    tosend = {Req: 0};
        //send the request
    $.post(host+"/initcluster",JSON.stringify(tosend)).done(function(data) {
        //update the peer list...
        console.log("OK for search");
    });


}

function SendBroadcast(){
    let content = $("#broadcastcontent").val();
    let destination = ""
    console.log("Sent " + content + " as broadcast");
    tosend = {"destination" : destination, "content" : content};
    $.post(host+"/broadcastmsg",JSON.stringify(tosend)).done(function(data){
        //update the peer list...
        console.log("Broadcast message got response : " + data)
        $("#clustermembers").empty();
        res = JSON.parse(data);

        for(let i = 0 ; i < res.length ; i ++){

            $("#clustermembers").append("<li id='member'>"+res[i]+"</li>");

        }
        $("li[id=member]").click(showanonymous);


    });
    //closing
    $("#broadcastcontent").val(" ");

    $("#clusterpopup").hide();


}

function showanonymous(){
    console.log("showing anonymous pannel..."); 
}