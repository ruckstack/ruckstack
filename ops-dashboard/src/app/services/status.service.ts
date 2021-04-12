import {Injectable, OnInit} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {UserService} from "./user.service";
import {BehaviorSubject, Observable, timer} from "rxjs";
import {StatusDetailedModel, StatusModel, Tracker} from "../models/status.model";
import {filter} from "rxjs/operators";

@Injectable({
  providedIn: 'root'
})
export class StatusService {

  private knowStatus = false;

  private status = new BehaviorSubject<StatusModel>({
    healthy: false,
    buildTime: 0,
    name: "",
    support: [],
    version: "",
    buildDate: new Date(),
  });

  constructor(private http: HttpClient, private userService: UserService) {
    this.checkStatus()

    setInterval(() => {
      this.checkStatus()
    }, 5000)
  }

  public get status$(): Observable<StatusModel> {
    return this.status.pipe(
      filter(data => this.knowStatus)
    );
  }

  private checkStatus() {
    if (this.userService.isLoggedIn()) {
      this.http.get<StatusDetailedModel>("/ops/api/status/detailed").subscribe((data: StatusDetailedModel) => {
        data.buildDate = new Date(data.buildTime * 1000);
        this.knowStatus = true;
        this.status.next(data);
      })
    } else {
      this.http.get<StatusModel>("/ops/api/status").subscribe((data: StatusModel) => {
        data.buildDate = new Date(data.buildTime * 100);
        this.knowStatus = true;
        this.status.next(data);
      })
    }
  }
}
