import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { Navbar } from '../navbar/navbar';

/** Authenticated shell: navbar + routed page content. */
@Component({
  selector: 'cu-main-layout',
  imports: [RouterOutlet, Navbar],
  template: `
    <cu-navbar />
    <main class="cu-container cu-page">
      <router-outlet />
    </main>
  `,
})
export class MainLayout {}
