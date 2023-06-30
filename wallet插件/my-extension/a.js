document.addEventListener("DOMContentLoaded", function () {
  var random = document.getElementById("random");
  random.addEventListener("click", function () {
    $.ajax({
      url: "http://127.0.0.1:8080/wallet",
      type: "POST",
      success: function (response) {
        $("#inputPublic").val(response["public_key"]);
        $("#inputPrivateKey").val(response["private_key"]);
        $("#inputAddress").val(response["blockchain_address"]);
        console.info(response);
      },
      error: function (error) {
        console.error(error);
      },
    });
  });

  var refresh = document.getElementById("refresh");
  refresh.addEventListener("click", function () {
    getTransactions();
  });

  var load_privatekey = document.getElementById("load_privatekey");
  load_privatekey.addEventListener("click", function () {
    let privatekey = $("#inputPrivateKey").val();
    console.log(privatekey);
    $.ajax({
      url: "http://127.0.0.1:8080/walletByPrivatekey",
      type: "POST",
      data: {
        privatekey: privatekey,
      },
      success: function (response) {
        $("#inputPublic").val(response["public_key"]);
        $("#inputPrivateKey").val(response["private_key"]);
        $("#inputAddress").val(response["blockchain_address"]);
        console.info(response);
      },
      error: function (error) {
        console.error(error);
      },
    });
  });

  $("#buttonSubmit").on("click", function () {
    var receiveAddress = $("#inputReceiveAddress").val();
    var amount = $("#inputAmount").val();

    if (receiveAddress.trim() === "") {
      alert("接收者账户地址不能为空");
      return;
    }

    if (isNaN(amount) || amount <= 0) {
      alert("金额必须为大于零的数字");
      return;
    }

    let public_key = $("#inputPublic").val();
    let private_key = $("#inputPrivateKey").val();
    let blockchain_address = $("#inputAddress").val();

    if (private_key.trim() === "") {
      alert("发送者私钥不能为空");
      return;
    }

    if (public_key.trim() === "") {
      alert("发送者公钥不能为空");
      return;
    }

    if (blockchain_address.trim() === "") {
      alert("发送者账户地址不能为空");
      return;
    }

    $.ajax({
      url: "http://127.0.0.1:8080/transaction",
      type: "POST",
      data: JSON.stringify({
        sender_public_key: public_key,
        sender_private_key: private_key,
        sender_blockchain_address: blockchain_address,
        recipient_blockchain_address: receiveAddress,
        value: amount,
      }),
      success: function (response) {
        if (response["message"] == "success") {
          alert("转账成功");
        } else {
          alert("转账失败");
        }
      },
      error: function (error) {
        console.error(error);
      },
    });
  });
});

function getTransactions() {
  $.ajax({
    url: "http://127.0.0.1:5000/getTransactions",
    type: "GET",
    dataType: "text",
    success: function (response) {
      let jsonArray = JSON.parse(response);

      const tableBody = $("#myTable tbody");

      // 清空表格内容
      tableBody.empty();

      // 遍历 JSON 数组
      $.each(jsonArray, function (index, item) {
        const row = $("<tr>");
        const sender_address = $("<td>").text(item.sender_blockchain_address);
        const recipient_address = $("<td>").text(
          item.recipient_blockchain_address
        );
        const value = $("<td>").text(item.value);
        const hash = $("<td>").text(item.hash);
        row.append(sender_address);
        row.append(recipient_address);
        row.append(value);
        row.append(hash);

        tableBody.append(row);
      });
    },
    error: function (error) {
      console.log(error);
    },
  });
}

$(document).ready(function () {
  getTransactions();
});
