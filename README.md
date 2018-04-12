ec2_recordsetter
====
 
EC2インスタンス起動時にRoute53へAレコードを追加する。

## Requirement
- 定義済みのHostedZone（プライベートゾーン推奨）
- Lambda用IAMロール
    - EC2
    - Route53
    - CloudWatchLogs

## Usage
1．zip状態のバイナリをLambdaにアップロードする。  
2．トリガーをCloudWatchEventsにして、以下のような設定にする。
````
{
  "source": [
    "aws.ec2"
  ],
  "detail-type": [
    "EC2 Instance State-change Notification"
  ],
  "detail": {
    "state": [
      "pending"
    ]
  }
}
````
3．LambdaにIAMロールを関連づける。  
4．環境変数に以下を追加する。  
````
DOMAIN     # 定義済みのドメイン名
HOSTZONE   # HostedZoneのID
REGION     # EC2インスタンスのリージョン
TAGKEY     # ホスト名として利用するタグ名
TTL        # TTLの秒数(特にコダワリがない場合は600)
````
5．設定を保存する。

## Install
ソースからコンパイルする場合、LambdaはLinuxで動く為、ビルドオプションを忘れないようにする。
````
GOOS=linux go build
````