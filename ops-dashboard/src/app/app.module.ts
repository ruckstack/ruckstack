import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { LayoutModule } from '@angular/cdk/layout';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { KubeDashboardComponent } from './kube-dashboard/kube-dashboard.component';
import { TraefikDashboardComponent } from './traefik-dashboard/traefik-dashboard.component';
import { HomeComponent } from './dashboard/home/home.component';
import {MatGridListModule} from "@angular/material/grid-list";
import {MatCardModule} from "@angular/material/card";
import {MatTableModule} from "@angular/material/table";
import {HttpClientModule} from "@angular/common/http";
import { StatusComponent } from './status/status.component';
import {MatMenuModule} from "@angular/material/menu";
import { UnauthorizedComponent } from './unauthorized/unauthorized.component';

@NgModule({
  declarations: [
    AppComponent,
    KubeDashboardComponent,
    TraefikDashboardComponent,
    HomeComponent,
    StatusComponent,
    UnauthorizedComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
    BrowserAnimationsModule,
    LayoutModule,
    MatGridListModule,
    MatCardModule,
    MatToolbarModule,
    MatButtonModule,
    MatSidenavModule,
    MatIconModule,
    MatListModule,
    MatTableModule,
    MatMenuModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
