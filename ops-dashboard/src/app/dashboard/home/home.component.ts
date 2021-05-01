import {Component, OnDestroy, OnInit} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {StatusDetailedModel, StatusModel, Tracker} from "../../models/status.model";
import {BehaviorSubject, Observable} from "rxjs";
import {CanActivate, CanDeactivate} from "@angular/router";
import {StatusService} from "../../services/status.service";
import {tap} from "rxjs/operators";

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnInit {

  healthy = false
  trackers: Tracker[] = []
  supportInfo: string[] = [];
  productName: string = "";
  buildDate: Date = new Date();
  version: string = "";

  currentProblems: string[] = []
  currentWarnings: string[] = []
  currentHealthy: string[] = []

  constructor(private statusService: StatusService) {
  }

  ngOnInit(): void {
    this.statusService.status$.pipe(
      tap(model => {
        this.healthy = model.healthy;
        this.trackers = (model as StatusDetailedModel).trackers;
        this.supportInfo = model.support;
        this.productName = model.name;
        this.version = model.version;
        this.buildDate = model.buildDate;

        this.currentProblems = [];
        this.currentWarnings = [];
        this.currentHealthy = [];

        if (this.trackers != null) {
          for (let tracker of this.trackers) {
            tracker.CurrentProblems.forEach((value, key) => {
              this.currentProblems = this.currentProblems.concat(key + ": " + value)
            })
            tracker.CurrentWarnings.forEach((value, key) => {
              this.currentWarnings = this.currentWarnings.concat(key + ": " + value)
            })
            if (tracker.CurrentProblems.size == 0 && tracker.CurrentWarnings.size == 0) {
              this.currentHealthy = this.currentHealthy.concat(tracker.Name)
            }
          }
        }
      })
    ).subscribe()

    this.statusService.checkStatus()
  }
}
