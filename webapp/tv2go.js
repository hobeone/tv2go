angular.module('tv2go', [
  'ngAnimate',
  'ui.router',
  'ng.group',
  'xeditable',
  'shows',
  'shows.episodes',
])
.config(function($stateProvider, $urlRouterProvider){
  'use strict';
  $urlRouterProvider.otherwise('/shows');

  $stateProvider
  .state('tv2go', {
    templateUrl: 'layout.tmpl.html',
    controller: 'IndexCtrl as indexCtrl',
    abstract: true,
  });
})
.run(function(editableOptions) {
  editableOptions.theme = 'bs3';
})
.controller('IndexCtrl', function(){});
