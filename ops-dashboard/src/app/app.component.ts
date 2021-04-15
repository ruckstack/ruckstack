import {Component} from '@angular/core';
import {UserService} from "./services/user.service";
import {Observable} from "rxjs";
import {StatusService} from "./services/status.service";
import {tap} from "rxjs/operators";
import {Title} from "@angular/platform-browser";

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {

  productName: string;
  username: Observable<string>

  constructor(private statusService: StatusService, private userService: UserService, private titleService: Title) {
    this.productName = "Ops Dashboard"
    this.statusService.status$.pipe(
      tap(data => {
        if (data.name) {
          this.productName = data.name;
          this.titleService.setTitle(this.productName+" Ops")
        }
      })
    ).subscribe()
    this.username = userService.username$
  }

  doLogin(): void {
    this.userService.login()
  }

  doLogout(): void {
    this.userService.logout()
  }
}
