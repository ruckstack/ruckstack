import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import {KubeDashboardComponent} from "./kube-dashboard/kube-dashboard.component";
import {TraefikDashboardComponent} from "./traefik-dashboard/traefik-dashboard.component";
import {HomeComponent} from "./dashboard/home/home.component";
import {StatusComponent} from "./status/status.component";
import {UserService} from "./services/user.service";
import {UnauthorizedComponent} from "./unauthorized/unauthorized.component";

const routes: Routes = [
  { path: 'status', component: StatusComponent },
  { path: 'unauthorized', component: UnauthorizedComponent },

  { path: 'dashboard', component: HomeComponent, canActivate: [UserService] },
  { path: 'dashboard/kubernetes', component: KubeDashboardComponent, canActivate: [UserService] },
  { path: 'dashboard/traefik', component: TraefikDashboardComponent, canActivate: [UserService] },
  { path: '**', redirectTo: "status" },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
