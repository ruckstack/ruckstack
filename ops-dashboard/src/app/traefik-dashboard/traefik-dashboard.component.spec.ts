import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TraefikDashboardComponent } from './traefik-dashboard.component';

describe('TraefikDashboardComponent', () => {
  let component: TraefikDashboardComponent;
  let fixture: ComponentFixture<TraefikDashboardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TraefikDashboardComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(TraefikDashboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
