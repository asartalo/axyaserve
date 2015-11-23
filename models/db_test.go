package models_test

import (
	"os"

	"github.com/asartalo/axyaserve/models"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
)

var sharedConnInfo = "./test.db"

func dropTestDb() {
	os.Remove(sharedConnInfo)
}

var _ = Describe("DB", func() {

	Context("Given there is no db yet", func() {
		BeforeEach(func() {
			dropTestDb()
		})

		Context("When a db is created", func() {
			var appDb models.AppDb
			var err error

			BeforeEach(func() {
				appDb, err = models.CreateDb(sharedConnInfo)
			})

			AfterEach(func() {
				appDb.Close()
			})

			It("There should be no errors", func() {
				Expect(err).To(BeNil())
			})

			Context("And a user is created", func() {
				var newuser models.User
				var newerr error
				var db *sqlx.DB

				BeforeEach(func() {
					newuser, newerr = appDb.NewUser("johndoe", "secret")
					db = appDb.DbStore()
				})

				findUser := func() (models.User, error) {
					user := models.User{}
					return user, db.Get(&user, "SELECT * FROM user WHERE name=?", "johndoe")
				}

				Context("When we check database", func() {
					var user models.User
					var err error

					BeforeEach(func() {
						user, err = findUser()
					})

					It("Saves user to db", func() {
						Expect(err).To(BeNil())
						Expect(user.Name).To(Equal("johndoe"))
					})

					It("Should have hashed password on user", func() {
						passerr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret"))
						Expect(passerr).To(BeNil())
					})
				})

				It("Returns new user and no error", func() {
					user, _ := findUser()
					Expect(newerr).To(BeNil())
					Expect(newuser).To(Equal(user))
				})

				It("Authenticates User", func() {
					user, err := appDb.Authenticate("johndoe", "secret")
					Expect(err).To(BeNil())
					Expect(user.Name).To(Equal("johndoe"))

					user, err = appDb.Authenticate("johndoe", "foo")
					Expect(err).To(Equal(models.UserAuthenticationError{"wrong password"}))
					Expect(user).To(Equal(models.User{}))
				})
			})

			Context("And attempting to authenticate a non-existing user", func() {
				var user models.User
				var err error
				BeforeEach(func() {
					user, err = appDb.Authenticate("foo", "bar")
				})

				It("Should not be authenticated", func() {
					Expect(err).To(Equal(models.UserAuthenticationError{"user not found"}))
					Expect(user).To(Equal(models.User{}))
				})
			})
		})

	})
})
