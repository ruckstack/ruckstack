import {Component, OnDestroy, OnInit} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {StatusDetailedModel, StatusModel, Tracker} from "../models/status.model";
import {CanDeactivate, NavigationStart, Router} from "@angular/router";
import {StatusService} from "../services/status.service";
import {Observable} from "rxjs";
import {tap} from "rxjs/operators";

@Component({
  selector: 'app-status',
  templateUrl: './status.component.html',
  styleUrls: ['./status.component.scss']
})
export class StatusComponent implements OnInit {

  healthy = false
  supportInfo: string[] = [];
  productName: string = "";
  buildDate: Date = new Date();
  version: string = "";

  constructor(public statusService: StatusService) {
  }

  ngOnInit(): void {
    this.statusService.status$.pipe(
      tap(model => {
        this.healthy = model.healthy;
        this.supportInfo = model.support;
        this.productName = model.name;
        this.version = model.version;
        this.buildDate = model.buildDate;
      })
    ).subscribe()
  }
}
