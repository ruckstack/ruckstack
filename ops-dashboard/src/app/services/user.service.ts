import {Injectable, OnInit} from '@angular/core';
import {BehaviorSubject, Observable} from "rxjs";
import {UserModel} from "../models/user.model";
import {HttpClient} from "@angular/common/http";
import {CanActivate, Router} from "@angular/router";
import {filter, map, withLatestFrom} from "rxjs/operators";

@Injectable({
  providedIn: 'root'
})
export class UserService implements CanActivate {

  private username = new BehaviorSubject<string>("")
  private ready = new BehaviorSubject<boolean>(false)

  constructor(private http: HttpClient, private router: Router) {
    this.http.get<UserModel>("/ops/api/me").subscribe((data: UserModel) => {
      this.username.next(data.username);
      this.ready.next(true)
    })
  }

  canActivate(): Observable<boolean> {
    return this.ready.pipe(
      filter(isReady => !!isReady),
      withLatestFrom(this.username),
      map(([ready, username]) => {
          if (!this.username.getValue()) {
            this.router.navigateByUrl("/unauthorized")
            return false;
          } else {
            return true;
          }
        }
      )
    )
  }

  public isLoggedIn(): boolean {
    return !!this.username.getValue();
  }

  public get username$(): Observable<string> {
    return this.username.asObservable();
  }

  public login() {
    this.http.get<UserModel>("/ops/api/login").subscribe((data: UserModel) => {
      this.username.next(data.username);
      this.router.navigate(["/dashboard"])
    })

  }

  public logout() {
    this.username.next("");
    this.router.navigate(["/ops"])
    this.http.get("/ops/api/logout").subscribe((data) => {
        console.log("Logout successful")
      },
      error => {
        console.log("logout error");
        console.log(error);
      })
  }
}
