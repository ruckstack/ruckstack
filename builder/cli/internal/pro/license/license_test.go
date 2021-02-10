package license

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		licenseText string
	}
	tests := []struct {
		name    string
		args    args
		want    *License
		wantErr string
	}{
		{
			name: "Valid license",
			args: args{
				licenseText: "01BkRhfgFN4nkhJmVkJBcSJj2Ampzo1Mj3p4LphZ/aFXIvCrOWK3HGzBy0LlzDCuY/4FFRAzD6h2+yIl3ko8CZMZbkJx3Q4zYeHewZL3dhY7Q44135iTeMsakmNftsHA7x1Ajk7tI30hN0y56NEsQZ99tzucueXYdAORnQ86jY3Lg8IR98CoGWGY1Ei55Qz1WNO8DWiBQBVBGoXZfaYYirsJVZXZ2GayHTtYtXPzHyJTUcurL5cJOfwWRIyHUZV5XeKwWioa30letG0l4x15mGnUVRDYho2g6w9nZQIUDPA692KbPUjtQmnTlATIRUugKC9OaGRv3+QgtV3rjiS+i7BxYP71UIEc5zynlU7lXL+LMvH1FBUMB2LZxsF/jcePmEQl/SnIy09XvTsxnF7qryxltF8zoAXJRvuOwdoYu7WNbBLDvd0/nbQj1Y9ZBMAnzfjWNesE7SkGu8FPXi3ew9EF2dJY2wZLB0sj+oK3KhXuruT/1p8JX5akmn0biVZg3I3aW9dPP+H6Zflb8HhH9cKPXylH/b6+XABaVRtbwqx1D7o5UTLNeNk+T+S/squqSjaU3wt6t0G1ThmfcrClMlV/1zWY3NewyQoTcCb+P/MENBaiTwF9e17Ys+9Xv7pbiAZaiVvO0GSAbyddL3C2S95dOZ3B+8WiQTa3MKoHA4FRk=-bmFtZTogVGVzdCBVc2VyCmVtYWlsOiB0ZXN0QGV4YW1wbGUuY29tCmV4cGlyZXM6IDE2NDQwMjgyOTgK",
			},
			want: &License{
				Name:    "Test User",
				Email:   "test@example.com",
				Expires: 1644028298,
			},
		},
		{
			name: "Invalid signature",
			args: args{
				licenseText: "01aW52YWxpZA==-bmFtZTogVGVzdCBVc2VyCmVtYWlsOiB0ZXN0QGV4YW1wbGUuY29tCmV4cGlyZXM6IDE2NDQwMjgyOTgK",
			},
			wantErr: "invalid license: verification failed",
		},
		{
			name: "Corrupted license",
			args: args{
				licenseText: "01BkRhfgFN4nkhJmVkJBcSJj2Ampzo1Mj3p4LphZ/aFXIvCrOWK3HGzBy0LlzDCuY/4FFRAzD6h2+yIl3ko8CZMZbkJx3Q4zYeHewZL3dhY7Q44135iTeMsakmNftsHA7x1Ajk7tI30hN0y56NEsQZ99tzucueXYdAORnQ86jY3Lg8IR98CoGWGY1Ei55Qz1WNO8DWiBQBVBGoXZfaYYirsJVZXZ2GayHTtYtXPzHyJTUcurL5cJOfwWRIyHUZV5XeKwWioa30letG0l4x15mGnUVRDYho2g6w9nZQIUDPA692KbPUjtQmnTlATIRUugKC9OaGRv3+QgtV3rjiS+i7BxYP71UIEc5zynlU7lXL+LMvH1FBUMB2LZxsF/jcePmEQl/SnIy09XvTsxnF7qryxltF8zoAXJRvuOwdoYu7WNbBLDvd0/nbQj1Y9ZBMAnzfjWNesE7SkGu8FPXi3ew9EF2dJY2wZLB0sj+oK3KhXuruT/1p8JX5akmn0biVZg3I3aW9dPP+H6Zflb8HhH9cKPXylH/b6+XABaVRtbwqx1D7o5UTLNeNk+T+S/squqSjaU3wt6t0G1ThmfcrClMlV/1zWY3NewyQoTcCb+P/MENBaiTwF9e17Ys+9Xv7pbiAZaiVvO0GSAbyddL3C2S95dOZ3B+8WiQTa3MKoHA4FRk=-bmFtZTogVGVzdCBVc2VyCmVtYWlsOiB0ZXN0QGV4YW1wbGUuY29tCmV4cGlyZXM6IDE2NDQwMjgyOTg",
			},
			wantErr: "invalid license: data corrupted",
		},
		{
			name: "Expired",
			args: args{
				licenseText: "01ydeMs/QuIajrBqDKaom0Ok1ZCgS6jX42TAo8Ybdi6R4MKiMiQohZ0zjaN9VpnbpZcnaJwrhET7xrkMA5ijYVFUo0J0MrYKHc9XlR2DMs96uKNql5OrfffDg4X1ovC1Fz1CF3rvK12mtERwzI5OTpNMbmN1s9F/mFR4+Oxilc4MPdxKG/St1m/dIIGc3KtoK4Q7lYXXqCfiWnimReYeRLco05u+i65P6c2tJuGNq6Uz6cVVugC9+g5xHpP2Mb3LgQFe+xS+KCrm/jZGoBz1K89na05ZTA1ciRoT4T0Yg+kWZWH8J5q6qWUUuSuC2/0/RSYc5SQRpdGiE1pEbkWw5mgu6rIfSwcttY0WLY+iMYoO7zGGXljNWs1KHHgPS9RqxZnMLVZMD/yXBOcjWzFDJpayBPesvgYcTTCDbDeS3+ONxed52uLPM7q6yfOgyuXVWVAv09r6Jkh+iMZRzW/R+qAFKUIZjJSIozeJeI0u/dP3756Mb3NQ5mpAehCmplZGmlIrFoQ4kcyjhFBWDshNDDmtr+7blcsPc1eARY9wUibYWJtVMYPXPYNwoB8bltlP3HdMCgXiK4tUd1pRoFcebGtkTqOp6maq8Lj1axeCXYK3eFsNGrTN8CA+6UHEyA/Y4Qd/gveeXz9hPJDG0QncTvlESnd1bdFejLG/hdKgEEyJ8=-bmFtZTogVGVzdCBVc2VyCmVtYWlsOiB0ZXN0QGV4YW1wbGUuY29tCmV4cGlyZXM6IDE2MTI4NDE0MzUK",
			},
			wantErr: "invalid license: expired",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse(tt.args.licenseText)
			if tt.wantErr == "" {
				if assert.NoError(t, err) {
					assert.NotNil(t, got)
					assert.Equal(t, tt.want.Name, got.Name)
					assert.Equal(t, tt.want.Email, got.Email)
					assert.Equal(t, tt.want.Expires, got.Expires)
				}
			} else {
				if assert.Error(t, err, tt.wantErr) {
					assert.Equal(t, tt.wantErr, err.Error())
				}
			}

		})
	}
}
