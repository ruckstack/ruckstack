import { Component, OnInit } from '@angular/core';
import {HttpClient} from "@angular/common/http";

@Component({
  selector: 'app-kube-dashboard',
  templateUrl: './kube-dashboard.component.html',
  styleUrls: ['./kube-dashboard.component.scss']
})
export class KubeDashboardComponent implements OnInit {

  constructor() { }

  ngOnInit(): void {
  }

}
