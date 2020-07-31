let user = new Vue({
    el: "#user",
    data: {
        id: "",
        username: "",
        teacher: 0,
        note: ""
    }
});

let userOnload = null;

function UserLogout() {
    Ajax(Api("logout"), null, null, "GET", function () {
        window.location.href = "../index/index.html"
    }, null);
}

function GetUser() {
    Ajax(Api("user/get"), null, null, "GET", function (data) {
        user.$data.id = data.Id;
        user.$data.username = data.Username;
        user.$data.teacher = data.Teacher;
        user.$data.note = data.Note;
        if (userOnload) {
            userOnload();
        }
    }, function (e) {
        if (e.status === 401) {
            alert("Login required.");
            window.location.href = "../user/login.html";
        } else {
            HandleError(e);
        }
    });
}

$(function () {
    GetUser();
})