import { ComponentFixture, TestBed } from '@angular/core/testing';

import { KubeDashboardComponent } from './kube-dashboard.component';

describe('KubeDashboardComponent', () => {
  let component: KubeDashboardComponent;
  let fixture: ComponentFixture<KubeDashboardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ KubeDashboardComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(KubeDashboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
