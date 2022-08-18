resource "aws_iam_user" "noirgate" {
  name = "noirgate-manager-user"
}

resource "aws_iam_access_key" "noirgate" {
  user = aws_iam_user.noirgate.name
}

resource "aws_iam_user_policy" "noirgate" {
  name = "noirgate-manager-policy"
  user = aws_iam_user.noirgate.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:CreateBucket",
        "s3:DeleteBucket",
        "s3:DeleteObject",
        "s3:GetBucketLocation",
        "s3:PutBucketAcl",
        "s3:PutBucketPolicy",
        "s3:PutBucketTagging",
        "s3:PutBucketCors",
        "s3:PutBucketWebsite"
      ],
      "Effect": "Allow",
      "Resource": ["arn:aws:s3:::noirgate-*"]
    }
  ]
}
EOF
}

