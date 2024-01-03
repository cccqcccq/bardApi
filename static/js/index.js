// 聊天框dom
var TalkDom = $(".talk_area ul")
// 聊天框
let textarea = $("textarea")
// 聊天记录
let chatHistory = {
    contents:[
        // 默认构建一个多轮对话,没有对话时bard返回错误Please ensure that multiturn requests alternate between user and model,正确的处理方式是先调用bard接口,获取一次聊天记录后通过more接口调用,由于不想写数据库所以直接投机取巧
        {role:"user",parts:[{text:"Forget everything that comes ahead"}]},
        {role:"model",parts:[{text:"Ok,I will forget everything that comes ahead"}]}
    ]
}
textarea.on('keydown',function(e){
    if(e.key === "Enter"){
        if(textarea.val() === ""){
            alert("发送消息不能为空")
            return
        }
        // 将用户输入的内容添加到聊天记录中
        let userChat = {role:"user",parts:[{text:textarea.val()}]}
        // 增加页面上的聊天内容
        $(".small_txt").text(textarea.val())
        chatHistory.contents.push(userChat)
        $(".talk_area ul").append(`<li class="right"><div class="content"><p>` + textarea.val() + `</p></div><span class="img"><img src="static/images/default2.jpg"/></span></li>`)
        $(".talk_area ul").append(`<li class="left"><span class="img"><img src="static/images/default.jpg"/></span><div class="content"><p class="text">正在思考中...</p></div></li>`)
        let  theText = $(".text:last")
        $.ajax({
            url:"/bard-more",
            type:"post",
            contentType:"application/json",
            data:JSON.stringify(chatHistory),
            success:function(e){
                theText.text(e.text)
                $(".small_txt").text(e.text)
                // 将bard的回复加入聊天记录中
                let modelChat = {role:"model",parts:[{text:e.text}]}
                chatHistory.contents.push(modelChat)
                console.log(chatHistory)
            }
        })
        // 清空输入框
        textarea.val('')
        // 聊天框滚动至底部
        ScrollToBottom()
        return false
    }
});

// 聊天框滚动至底部
function ScrollToBottom(){
    TalkDom.scrollTop(TalkDom.prop("scrollHeight"));
}

// 获取当前时分
function getNowTime(){
    let myDate = new Date();
    let h = myDate.getHours(); // 获取当前小时数(0-23)
    let m = myDate.getMinutes(); // 获取当前分钟数(0-59)
    h = h < 10?"0"+h:h
    m = m < 10?"0"+m:m
    return h+":"+m
}


