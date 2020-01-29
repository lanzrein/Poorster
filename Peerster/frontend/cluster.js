/**
 * @file cluster.js contains the handler to connect to the server and make requests for the cluster management
 //@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent
 */

function InitCluster(){

    console.log(host);
    tosend = {Req: 0};
        //send the request
    $.post(host+"/initcluster",JSON.stringify(tosend)).done(function(data) {
        console.log("OK for init");
    });


}


function LeaveCluster(){
    console.log(host);
    tosend = {Req: 0};
    //send the request
    $.post(host+"/leavecluster",JSON.stringify(tosend)).done(function(data) {
        console.log("OK for leave");
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


let anonFlag = false;

function anonmessage(){
    dst = $(this).parent().text();
    if (dst.length > 5) {
        dst = dst.substring(0,dst.length-5);
    }
    console.log("anon message ! "+ dst ) ;

    $("#receiver").text(dst);
    $("#privatepopup").show();
    $("#clusterpopup").hide();
    $("#anonparams").show();
    anonFlag = true ;
}

let dst;
function anoncall(){
    dst = $(this).parent().text() ;
    console.log(dst.length);
    if (dst.length >= 5) {
        dst = dst.substring(0,dst.length-4);
    }else{
        alert("Incorrect name");
    }
    console.log("anon call to : " + dst   );
    $("#callee").text(dst);
    $("#callpannel").show();

}
let currently_calling;


function accept_call(){
    let callee = $("#callee").text()
    currently_calling = callee;
    val = { "Accept" : true, "Member":callee};
    post_call_data(val)
    console.log("Accept on :", JSON.stringify(val));
}

function decline_call(){
    val = { "Decline" : true, "Member": $("#callee").text()};
    console.log("hangup on :", JSON.stringify(val));
    post_call_data(val)
    close_call();
}

function hangup_call(){
    val = { "Hangup" : true, "Member": $("#callee").text()};
    console.log("Declining on :", JSON.stringify(val));
    post_call_data(val)
    close_call();
}

function dial_call(){
    val = { "Dial" : true, "Member": $("#callee").text()};
    console.log("Dial on :", JSON.stringify(val));
    post_call_data(val)
}

function post_call_data(tosend){
    $.post(host+"/callhandler",JSON.stringify(tosend)).done(function(data) {
        console.log("OK for posting"); ;
    });
}

function close_call(){
    $("#callee").text("");
    $("#callpannel").hide();
}

let person;
let vote;

function openvotepannel(){
    let text = $(this).text();
    let xs = text.split(" ");
    if (xs.length !== 2){
        console.log("Unexpected length of array");
        return ;
    }

    vote = xs[0];
    person = xs[1];

    console.log("Vote "+vote + "person "+person);
    $("#votepannel").show();
    $("#votetext").text("Do you want to "+vote+" person " + person +" to the cluster ?") ;



}
function castvote(){
    let decision = $(this).text();
    let bool = decision === 'YES' ;
    tosend = {'Vote' : vote, 'Person' : person , 'Decision' : bool}
    console.log("Sending " + JSON.stringify(tosend));

    $.post(host+"/evoting",JSON.stringify(tosend)).done(function(data){
        //update the peer list...
        console.log("Evoting got response : " + data)
        $("#votelist").empty();
        res = JSON.parse(data);

        for(let i = 0 ; i < res.length ; i ++){

            $("#votelist").append("<li id='vote'>"+res[i]+"</li>");

        }
        $("li[id=vote]").click(openvotepannel);

    });

    closevote()
}
function closevote(){
    $("#votepannel").hide();
    $("#votetext").text('');
    person = '';
    vote = '';


}



function openjoin(){
    $("#joinpannel").show();

}

function joinrequest(){
    let other = $("#joinOther").val();
    if (other==""){
        alert("Empty member");
    }


    tosend = {"joinOther":other};
    $.post(host+"/joinrequest",JSON.stringify(tosend)).done(function(data){
        //update the peer list...
        console.log("Anonymous call got response : " + data)

    });

    closejoin()
}

function closejoin(){
    $("#joinpannel").hide();
    $("#joinOther").text("");
}
