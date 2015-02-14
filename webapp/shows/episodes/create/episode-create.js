angular.module('shows.episodes.create',['tv2go.episodesService'])
.config(function($stateProvider){
  $stateProvider
  .state('tv2go.shows.episodes.create', {
    url: "/episodes/create",
    templateUrl: "shows/episodes/create/episode-create.tmpl.html",
    controller: "CreateEpisodeCtrl as createEpisodeCtrl"
  });
})
.controller('CreateEpisodeCtrl', function($state, $stateParams, EpisodesModel, Episode){
  var createEpisodeCtrl = this;

  function returnToEpisodes(){
    $state.go("tv2go.shows.episodes", {
      show: $stateParams.show
    });
  }

  function cancelCreating() {
    returnToEpisodes();
  }

  function createEpisode(episode) {
    EpisodesModel.createEpisode(episode);
    returnToEpisodes();
  }

  function resetForm() {
    createEpisodeCtrl.newEpisode = new Episode();
  }

  createEpisodeCtrl.cancelCreating = cancelCreating;
  createEpisodeCtrl.createEpisode = createEpisode;

  resetForm();
})
;
