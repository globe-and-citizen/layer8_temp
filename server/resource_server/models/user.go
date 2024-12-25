package models

type User struct {
	ID               uint   `gorm:"primaryKey; unique; autoIncrement; not null" json:"id"`
	Username         string `gorm:"column:username; unique; not null" json:"username"`
	Password         string `gorm:"column:password; not null" json:"password"`
	FirstName        string `gorm:"column:first_name; not null" json:"first_name"`
	LastName         string `gorm:"column:last_name; not null" json:"last_name"`
	Salt             string `gorm:"column:salt; not null" json:"salt"`
	EmailProof       []byte `gorm:"column:email_proof; not null" json:"email_proof"`
	VerificationCode string `gorm:"column:verification_code; not null" json:"verification_code"`
	ZkKeyPairId      uint   `gorm:"column:zk_key_pair_id; not null" json:"zk_key_pair_id"`
	PublicKey        []byte `gorm:"column:public_key; not null" json:"public_key"`
	IterationCount   int    `gorm:"column:iterationCount;" json:"iterationCount"`
}

func (User) TableName() string {
	return "users"
}
