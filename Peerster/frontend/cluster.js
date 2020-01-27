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


let anonmessage = false;
function anonmessage(){
    dst = $(this).parent().text();
    if (dst.length > 5) {
        dst = dst.substring(0,dst.length-5);
    }
    console.log("anon message ! "+ dst ) ;
    $("#receiver").text(dst);
    $("#privatepopup").show()
    anonmessage = true ;
}

let dst;
function anoncall(){
    dst = $(this).parent().text() ;
    if (dst.length > 5) {
        dst = dst.substring(0,dst.length-5);
    }
    console.log("anon call to : " + dst   );
    tosend = {"destination":dst, "content":content};
    $.post(host+"/anoncall",JSON.stringify(tosend)).done(function(data){
        //update the peer list...
        console.log("Anonymous call got response : " + data)

    });



}